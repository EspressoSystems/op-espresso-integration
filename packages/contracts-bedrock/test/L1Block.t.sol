// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing utilities
import { CommonTest } from "test/CommonTest.t.sol";

// Target contract
import { L1Block } from "src/L2/L1Block.sol";

contract L1BlockTest is CommonTest {
    L1Block lb;
    address depositor;
    bytes32 immutable NON_ZERO_HASH = keccak256(abi.encode(1));

    /// @dev Sets up the test suite.
    function setUp() public virtual override {
        super.setUp();
        lb = new L1Block();
        depositor = lb.DEPOSITOR_ACCOUNT();
        vm.prank(depositor);
        lb.setL1BlockValues(
            L1Block.L1BlockValues({
                number: uint64(1),
                timestamp: uint64(2),
                basefee: 3,
                hash: NON_ZERO_HASH,
                sequenceNumber: uint64(4),
                batcherHash: bytes32(0),
                l1FeeOverhead: 2,
                l1FeeScalar: 3,
                espresso: false,
                espressoL1ConfDepth: 0,
                justification: "0xc0"
            })
        );
    }

    /// @dev Tests that `setL1BlockValues` updates the values correctly.
    function testFuzz_updatesValues_succeeds(
        uint64 n,
        uint64 t,
        uint256 b,
        bytes32 h,
        uint64 s,
        bytes32 bt,
        uint256 fo,
        uint256 fs,
        bool e,
        uint64 cd
    )
        external
    {
        vm.prank(depositor);
        lb.setL1BlockValues(
            L1Block.L1BlockValues({
                number: n,
                timestamp: t,
                basefee: b,
                hash: h,
                sequenceNumber: s,
                batcherHash: bt,
                l1FeeOverhead: fo,
                l1FeeScalar: fs,
                espresso: e,
                espressoL1ConfDepth: cd,
                justification: "0xc0"
            })
        );
        assertEq(lb.number(), n);
        assertEq(lb.timestamp(), t);
        assertEq(lb.basefee(), b);
        assertEq(lb.hash(), h);
        assertEq(lb.sequenceNumber(), s);
        assertEq(lb.batcherHash(), bt);
        assertEq(lb.l1FeeOverhead(), fo);
        assertEq(lb.l1FeeScalar(), fs);
        assertEq(lb.espresso(), e);
        assertEq(lb.espressoL1ConfDepth(), cd);
    }

    /// @dev Tests that `number` returns the correct value.
    function test_number_succeeds() external {
        assertEq(lb.number(), uint64(1));
    }

    /// @dev Tests that `timestamp` returns the correct value.
    function test_timestamp_succeeds() external {
        assertEq(lb.timestamp(), uint64(2));
    }

    /// @dev Tests that `basefee` returns the correct value.
    function test_basefee_succeeds() external {
        assertEq(lb.basefee(), 3);
    }

    /// @dev Tests that `hash` returns the correct value.
    function test_hash_succeeds() external {
        assertEq(lb.hash(), NON_ZERO_HASH);
    }

    /// @dev Tests that `sequenceNumber` returns the correct value.
    function test_sequenceNumber_succeeds() external {
        assertEq(lb.sequenceNumber(), uint64(4));
    }

    /// @dev Tests that `setL1BlockValues` can set max values.
    function test_updateValues_succeeds() external {
        vm.prank(depositor);
        lb.setL1BlockValues(
            L1Block.L1BlockValues({
                number: type(uint64).max,
                timestamp: type(uint64).max,
                basefee: type(uint256).max,
                hash: keccak256(abi.encode(1)),
                sequenceNumber: type(uint64).max,
                batcherHash: bytes32(type(uint256).max),
                l1FeeOverhead: type(uint256).max,
                l1FeeScalar: type(uint256).max,
                espresso: true,
                espressoL1ConfDepth: type(uint64).max,
                justification: hex"c0"
            })
        );
    }
}
