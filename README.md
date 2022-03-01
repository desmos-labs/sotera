# Soteria
> In Greek mythology, Soteria (Ancient Greek: Σωτηρία) was the goddess or spirit (daimon) of safety and salvation, deliverance, and preservation from harm (not to be mistaken for Eleos).   [[Source](https://en.wikipedia.org/wiki/Soteria_(mythology))]


This script allows to fix the bug that some chains might have after running the on-chain migration to cosmos `v0.43` (now `v0.44`) described [here](https://github.com/cosmos/cosmos-sdk/issues/10712).

To do this, it performs the following operations: 
1. reads a genesis file which path is provided as the first argument;
2. for each vesting account:
   1. it gets the original vesting data stored inside the genesis
   2. queries all the `MsgDelegate` and `MsgUnbond` transactions 
   3. calls the `TrackDelegation` and `TrackUndelegation` accordingly
3. outputs the final `x/vesting` state that should be present on chain

The output state can then be fed to an on-chain upgrade that reads such state and stores the various account data on-chain, effectively fixing the bug.