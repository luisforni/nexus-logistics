
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts/security/ReentrancyGuard.sol";
import "@openzeppelin/contracts/security/Pausable.sol";

contract ShipmentTracker is AccessControl, ReentrancyGuard, Pausable {
    bytes32 public constant RECORDER_ROLE = keccak256("RECORDER_ROLE");

    error ZeroAdminAddress();
    error InvalidShipmentIdLength();
    error InvalidStatusLength();
    error NotesTooLong();
    error TooManyEventsForShipment();

    event ShipmentEventRecorded(
        string  indexed shipmentId,
        string  status,
        string  notes,
        address indexed recorder,
        uint256 timestamp
    );

    event RecorderGranted(address indexed account, address indexed by);
    event RecorderRevoked(address indexed account, address indexed by);

    struct Event {
        string  status;
        string  notes;
        address recorder;
        uint256 timestamp;
    }

    mapping(string => Event[]) private _events;

    uint256 public totalEvents;

    uint256 private constant MAX_STR_LEN = 256;

    uint256 private constant MAX_EVENTS_PER_SHIPMENT = 200;

    constructor(address admin) {
        if (admin == address(0)) revert ZeroAdminAddress();
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(RECORDER_ROLE, admin);
    }

    function recordEvent(
        string calldata shipmentId,
        string calldata status,
        string calldata notes
    ) external onlyRole(RECORDER_ROLE) nonReentrant whenNotPaused {
        if (bytes(shipmentId).length == 0 || bytes(shipmentId).length > 36)
            revert InvalidShipmentIdLength();
        if (bytes(status).length == 0 || bytes(status).length > 64)
            revert InvalidStatusLength();
        if (bytes(notes).length > MAX_STR_LEN)
            revert NotesTooLong();
        if (_events[shipmentId].length >= MAX_EVENTS_PER_SHIPMENT)
            revert TooManyEventsForShipment();

        _events[shipmentId].push(Event({
            status:    status,
            notes:     notes,
            recorder:  msg.sender,
            timestamp: block.timestamp
        }));

        unchecked { ++totalEvents; }

        emit ShipmentEventRecorded(shipmentId, status, notes, msg.sender, block.timestamp);
    }

    function getEvents(string calldata shipmentId)
        external
        view
        returns (Event[] memory)
    {
        return _events[shipmentId];
    }

    function getLatestEvent(string calldata shipmentId)
        external
        view
        returns (Event memory)
    {
        Event[] storage evts = _events[shipmentId];
        require(evts.length > 0, "no events for shipment");
        return evts[evts.length - 1];
    }

    function getEventCount(string calldata shipmentId)
        external
        view
        returns (uint256)
    {
        return _events[shipmentId].length;
    }

    function grantRecorder(address account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _grantRole(RECORDER_ROLE, account);
        emit RecorderGranted(account, msg.sender);
    }

    function revokeRecorder(address account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _revokeRole(RECORDER_ROLE, account);
        emit RecorderRevoked(account, msg.sender);
    }

    function pause() external onlyRole(DEFAULT_ADMIN_ROLE) { _pause(); }
    function unpause() external onlyRole(DEFAULT_ADMIN_ROLE) { _unpause(); }
}
