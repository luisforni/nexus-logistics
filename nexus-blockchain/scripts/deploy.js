const { ethers } = require("hardhat");
const fs = require("fs");

async function main() {
  const [deployer] = await ethers.getSigners();
  const network = await ethers.provider.getNetwork();

  console.log(`\n🚀 Deploying to ${network.name} (chainId: ${network.chainId})`);
  console.log(`   Deployer: ${deployer.address}`);
  console.log(`   Balance:  ${ethers.formatEther(await ethers.provider.getBalance(deployer.address))} ETH\n`);

  const ShipmentTracker = await ethers.getContractFactory("ShipmentTracker");
  const tracker = await ShipmentTracker.deploy(deployer.address);
  await tracker.waitForDeployment();
  const trackerAddr = await tracker.getAddress();
  console.log(`✓ ShipmentTracker deployed: ${trackerAddr}`);

  const treasury = process.env.TREASURY_ADDRESS || deployer.address;
  const SupplyChainToken = await ethers.getContractFactory("SupplyChainToken");
  const token = await SupplyChainToken.deploy(deployer.address, treasury);
  await token.waitForDeployment();
  const tokenAddr = await token.getAddress();
  console.log(`✓ SupplyChainToken deployed: ${tokenAddr}`);

  const deployment = {
    network: network.name,
    chainId: network.chainId.toString(),
    deployer: deployer.address,
    deployedAt: new Date().toISOString(),
    contracts: {
      ShipmentTracker: trackerAddr,
      SupplyChainToken: tokenAddr,
    },
  };

  const outPath = `./deployments/${network.name}.json`;
  fs.mkdirSync("./deployments", { recursive: true });
  fs.writeFileSync(outPath, JSON.stringify(deployment, null, 2));
  console.log(`\n📄 Deployment saved to ${outPath}`);

  if (network.name !== "hardhat" && network.name !== "localhost") {
    console.log("\nTo verify contracts on Etherscan:");
    console.log(`  npx hardhat verify --network ${network.name} ${trackerAddr} "${deployer.address}"`);
    console.log(`  npx hardhat verify --network ${network.name} ${tokenAddr} "${deployer.address}" "${treasury}"`);
  }
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
