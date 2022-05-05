const path = require("path");
const fs = require("fs");
const { expect } = require("chai");

// const c_tester = require("circom_tester").c;
const wasm_tester = require("circom_tester").wasm;

describe("TestZKInputs 3 nMaxVotes, 3 nLevels", function () {
	this.timeout(0);
	let circuitFile = "zkinputstest.circom";
	let circuitPath = path.join(__dirname, "", circuitFile);
	let circuit;

	const nMaxVotes = 3;
	const nLevels = 3;

	before( async() => {
		console.log("generating compiled files for first time");
		const circuitCode = `
			pragma circom 2.0.0;
			include "./node_modules/zkmultisig/circuits/src/zkmultisig.circom";
			component main {public [chainID, processID, censusRoot, receiptsRoot, nVotes, result, withReceipts]} = zkmultisig(${nMaxVotes}, ${nLevels});
		`;
		await fs.writeFileSync(circuitPath, circuitCode, "utf8");

		console.log("going to load tester");
		// circuit = await c_tester(circuitPath);
		circuit = await wasm_tester(circuitPath);
		console.log("tester loaded, now going to circuit.loadConstraints");
		await circuit.loadConstraints();
		console.log("n_constraints", circuit.constraints.length);
	});

	after( async() => {
		fs.unlinkSync(circuitPath); // comment in case of going to reuse compiled circuit
	});

	it("zkinputs_3_3_1_60.json", async () => {
		const filePath = path.resolve(__dirname, './zkinputs_3_3_1_60.json');
		const data = fs.readFileSync(filePath);
		const inputs = JSON.parse(data);
		// console.log("inputs\n", inputs);

		console.log("calculateWitness");
		const witness = await circuit.calculateWitness(inputs, true);
		await circuit.checkConstraints(witness);
	});
	it("zkinputs_3_3_3_60.json", async () => {
		const filePath = path.resolve(__dirname, './zkinputs_3_3_3_60.json');
		const data = fs.readFileSync(filePath);
		const inputs = JSON.parse(data);

		const witness = await circuit.calculateWitness(inputs, true);
		await circuit.checkConstraints(witness);
	});
});

describe("TestZKInputs 8 nMaxVotes, 4 nLevels", function () {
	this.timeout(0);
	let circuitFile = "zkinputstest.circom";
	let circuitPath = path.join(__dirname, "", circuitFile);
	let circuit;

	const nMaxVotes = 8;
	const nLevels = 4;

	before( async() => {
		console.log("generating compiled files for first time");
		const circuitCode = `
			pragma circom 2.0.0;
			include "./node_modules/zkmultisig/circuits/src/zkmultisig.circom";
			component main {public [chainID, processID, censusRoot, receiptsRoot, nVotes, result, withReceipts]} = zkmultisig(${nMaxVotes}, ${nLevels});
		`;
		await fs.writeFileSync(circuitPath, circuitCode, "utf8");

		console.log("going to load tester");
		// circuit = await c_tester(circuitPath);
		circuit = await wasm_tester(circuitPath);

		console.log("tester loaded, now going to circuit.loadConstraints");
		await circuit.loadConstraints();
		console.log("n_constraints", circuit.constraints.length);
	});

	after( async() => {
		fs.unlinkSync(circuitPath); // comment in case of going to reuse compiled circuit
	});

	it("zkinputs_8_4_3_60.json", async () => {
		const filePath = path.resolve(__dirname, 'zkinputs_8_4_3_60.json');
		const data = fs.readFileSync(filePath);
		const inputs = JSON.parse(data);

		// console.log("inputs\n", inputs);

		console.log("calculateWitness");
		const witness = await circuit.calculateWitness(inputs, true);
		await circuit.checkConstraints(witness);
	});
	it("zkinputs_8_4_5_60.json", async () => {
		const filePath = path.resolve(__dirname, 'zkinputs_8_4_5_60.json');
		const data = fs.readFileSync(filePath);
		const inputs = JSON.parse(data);

		const witness = await circuit.calculateWitness(inputs, true);
		await circuit.checkConstraints(witness);
	});
	it("zkinputs_8_4_7_60.json", async () => {
		const filePath = path.resolve(__dirname, 'zkinputs_8_4_7_60.json');
		const data = fs.readFileSync(filePath);
		const inputs = JSON.parse(data);

		const witness = await circuit.calculateWitness(inputs, true);
		await circuit.checkConstraints(witness);
	});
	it("zkinputs_8_4_8_60.json", async () => {
		const filePath = path.resolve(__dirname, 'zkinputs_8_4_8_60.json');
		const data = fs.readFileSync(filePath);
		const inputs = JSON.parse(data);

		const witness = await circuit.calculateWitness(inputs, true);
		await circuit.checkConstraints(witness);
	});
});
