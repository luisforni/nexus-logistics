const { expect } = require("chai");
const { ethers } = require("hardhat");
const { loadFixture } = require("@nomicfoundation/hardhat-toolbox/network-helpers");

describe("ShipmentTracker", function () {
  async function deployFixture() {
    const [admin, recorder, stranger] = await ethers.getSigners();
    const ShipmentTracker = await ethers.getContractFactory("ShipmentTracker");
    const tracker = await ShipmentTracker.deploy(admin.address);
    await tracker.waitForDeployment();
    return { tracker, admin, recorder, stranger };
  }

  describe("Deployment", function () {
    it("grants DEFAULT_ADMIN_ROLE to the admin", async function () {
      const { tracker, admin } = await loadFixture(deployFixture);
      expect(await tracker.hasRole(await tracker.DEFAULT_ADMIN_ROLE(), admin.address)).to.be.true;
    });

    it("grants RECORDER_ROLE to the admin", async function () {
      const { tracker, admin } = await loadFixture(deployFixture);
      const RECORDER_ROLE = await tracker.RECORDER_ROLE();
      expect(await tracker.hasRole(RECORDER_ROLE, admin.address)).to.be.true;
    });

    it("starts with zero total events", async function () {
      const { tracker } = await loadFixture(deployFixture);
      expect(await tracker.totalEvents()).to.equal(0);
    });
  });

  describe("recordEvent", function () {
    it("records an event and emits the event log", async function () {
      const { tracker, admin } = await loadFixture(deployFixture);
      const id = "550e8400-e29b-41d4-a716-446655440000";

      await expect(tracker.recordEvent(id, "PENDING", "genesis"))
        .to.emit(tracker, "ShipmentEventRecorded")
        .withArgs(id, "PENDING", "genesis", admin.address, anyValue);

      expect(await tracker.totalEvents()).to.equal(1);
    });

    it("allows multiple events for the same shipment", async function () {
      const { tracker } = await loadFixture(deployFixture);
      const id = "test-shipment-001";
      await tracker.recordEvent(id, "PENDING", "");
      await tracker.recordEvent(id, "PICKED_UP", "picked");
      await tracker.recordEvent(id, "IN_TRANSIT", "on the way");

      expect(await tracker.getEventCount(id)).to.equal(3);
    });

    it("reverts for a stranger without RECORDER_ROLE", async function () {
      const { tracker, stranger } = await loadFixture(deployFixture);
      await expect(
        tracker.connect(stranger).recordEvent("id", "PENDING", "")
      ).to.be.reverted;
    });

    it("reverts on empty shipmentId", async function () {
      const { tracker } = await loadFixture(deployFixture);
      await expect(tracker.recordEvent("", "PENDING", "")).to.be.revertedWith(
        "invalid shipmentId length"
      );
    });

    it("reverts on notes exceeding 256 chars", async function () {
      const { tracker } = await loadFixture(deployFixture);
      const longNotes = "x".repeat(257);
      await expect(
        tracker.recordEvent("some-id", "IN_TRANSIT", longNotes)
      ).to.be.revertedWith("notes too long");
    });
  });

  describe("getLatestEvent", function () {
    it("returns the most recent event", async function () {
      const { tracker } = await loadFixture(deployFixture);
      const id = "latest-test";
      await tracker.recordEvent(id, "PENDING", "");
      await tracker.recordEvent(id, "DELIVERED", "done");
      const evt = await tracker.getLatestEvent(id);
      expect(evt.status).to.equal("DELIVERED");
      expect(evt.notes).to.equal("done");
    });

    it("reverts if no events exist", async function () {
      const { tracker } = await loadFixture(deployFixture);
      await expect(
        tracker.getLatestEvent("nonexistent")
      ).to.be.revertedWith("no events for shipment");
    });
  });

  describe("Access Control", function () {
    it("admin can grant and revoke RECORDER_ROLE", async function () {
      const { tracker, admin, recorder } = await loadFixture(deployFixture);
      const RECORDER_ROLE = await tracker.RECORDER_ROLE();

      await tracker.grantRecorder(recorder.address);
      expect(await tracker.hasRole(RECORDER_ROLE, recorder.address)).to.be.true;

      await tracker.revokeRecorder(recorder.address);
      expect(await tracker.hasRole(RECORDER_ROLE, recorder.address)).to.be.false;
    });
  });

  describe("Pause", function () {
    it("prevents recording when paused", async function () {
      const { tracker } = await loadFixture(deployFixture);
      await tracker.pause();
      await expect(
        tracker.recordEvent("id", "PENDING", "")
      ).to.be.revertedWith("Pausable: paused");
    });
  });
});

function anyValue() { return true; }
