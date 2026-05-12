
Terra Classic USTC Staking System -Technical Implementation Proposal
Following Signal Proposal 12219 (USTC Staking System), which passed, this proposal defines the technical scope, architecture, and implementation plan for introducing a USTC staking system on Layer 1.

Executive Summary
Following the successful signal from Proposal 12219, this proposal outlines the design and implementation of a native USTC staking system on Terra Classic Layer 1.
The objective is to introduce staking-like utility for USTC holders without modifying the core validator consensus model. This will be achieved by implementing a new dedicated module (x/ustcstaking), enabling users to delegate USTC, earn rewards, and participate in the ecosystem , while preserving LUNC as the sole consensus staking asset.
This approach ensures:
Zero impact on consensus security
Clean modular design aligned with Cosmos SDK standards
Scalable foundation for future enhancements (oracle tuning, fee-based rewards, etc.)

Important: The module ships with enabled = false. Activation requires a separate governance proposal once real yield sources (MM2, protocol fees) are confirmed live. This safeguard prevents reward pool depletion before sustainable yield exists.

Motivation
Utility Expansion:
USTC currently lacks native Layer 1 utility. Introducing staking creates immediate on-chain engagement.
Supply Locking:
Encourages voluntary locking of circulating USTC, reducing sell pressure and improving market structure.
Ecosystem Growth:
Attracts users to interact directly with Layer 1 in a simple and familiar way (staking UX).
Validator Alignment:
Allows delegations to validators, strengthening community alignment without affecting voting power.
Future Revenue Integration:
Provides a framework that can later integrate protocol revenue (MM2, fees) into reward distribution.

Scope of Work
1. New Module: x/ustcstaking
   Implement a dedicated module to handle USTC staking logic, separate from the native staking module.
   Core features:
   USTC delegation to validators
   Undelegation with configurable unbonding period
   Redelegation between validators
   Reward accrual and claiming
   Validator commission on USTC rewards
   Important:
   This module does not modify:
   Validator set selection
   Voting power
   Consensus mechanics
   LUNC remains the only staking asset for consensus.
   USTC principal is not subject to slashing for consensus faults. Because USTC stake does not participate in consensus security, validator misbehavior on the LUNC side does not automatically reduce USTC principal. Any penalty model for USTC delegators (if desired in future phases) must be a separate governance-controlled parameter and should affect rewards before principal.

2. Core Functionality
   Delegation System
   Users delegate uusd (USTC) to validators
   Delegations tracked independently from LUNC staking
   Validator serves as accounting endpoint only

Unbonding Logic
Configurable unbonding period (governance parameter)
Optional instant unbond: user may exit immediately by paying a penalty fee (instant_unbond_fee_bps, disabled initially)
Reward Distribution
Rewards distributed proportionally based on USTC stake
Epoch-based distribution; only validators in Bonded status at epoch start are eligible
Independent reward pool
Reward Pool
Funded via MsgFundUSTCRewardPool (governance-only in Phase 1; a fund_pool_admins parameter may whitelist additional addresses in future phases via governance)
When pool balance is insufficient, reward emission stops for that epoch; principal is not affected and a pool_exhausted event is emitted so delegators and operators are notified

3. Validator Status Handling
   The module reads validator status from x/staking (read-only) at EndBlock and applies the following rules:

Jailed validator: new USTC delegations are blocked; reward accrual stops at the start of the next epoch (the epoch in which jailing occurs is still distributed normally; rewards are not retroactively removed). In Phase 1, rewards are paused only — redirection to other delegators is not implemented. No retroactive catch-up after unjailing.
Unjailed validator: reward accrual resumes from the next epoch start forward. No retroactive catch-up for the jailed period.
Tombstoned / permanently removed validator: all USTC delegations to that validator enter forced unbonding immediately, using a governance-controlled forced_unbond_time parameter. Users may also redelegate out before unbonding completes.
Unbonding / Unbonded validator (not jailed): reward accrual pauses; new delegations are blocked. Existing delegations remain; users may redelegate or wait.
Redelegation in transit to a tombstoned validator: the redelegation is cancelled and funds are returned to the source validator delegation.

These rules ensure that USTC delegators are not penalised for consensus faults while preventing rewards from accruing to inactive or misbehaving validators.
4. Parameters (Governance Controlled)
   Initial parameters include:
   enabled
   bond_denom = uusd
   unbonding_time
   apr_bps (APR % Rewards)
   reward_epoch
   max_total_staked
   max_validator_share
   min_delegation
   instant_unbond_fee_bps (disabled initially)
   forced_unbond_time (applies when a validator is tombstoned)
   fund_pool_admins (initially empty; governance may whitelist addresses to call MsgFundUSTCRewardPool)
   oracle_enabled (disabled in Phase 1)
   All parameters adjustable via governance.

5. Message Types
   MsgDelegateUSTC
   MsgUndelegateUSTC
   MsgBeginRedelegateUSTC
   MsgWithdrawUSTCRewards
   MsgFundUSTCRewardPool
   MsgUpdateUSTCParams
   MsgEmergencyPause — halts new delegations and reward distribution immediately; unbonding and redelegation out remain available. Callable by governance only. Intended for critical incidents (e.g. discovered exploit, pool manipulation). Resumption requires a separate governance proposal.

6. Reward Model (Phase 1)
   Fixed Reward Model (Initial Implementation):
   Rewards emitted from a pre-funded pool
   APR set conservatively via governance
   Distribution proportional to stake
   This ensures:
   Simplicity
   Predictability
   Controlled rollout

7. Oracle Mechanism (Future Phase)
   A separate lightweight oracle system may be introduced in later phases to:
   Suggest APR adjustments
   Monitor staking participation
   Provide validator-submitted signals
   This will:
   NOT interfere with Terra’s existing price oracle
   Be bounded by governance-defined limits

8. Integration
   Validators
   Act as delegation endpoints
   Receive commission on USTC rewards at the rate set in x/staking (read live at each reward distribution; no separate USTC commission parameter in Phase 1)
   No impact on consensus power
   Wallets & dApps
   Not required for Phase 1
   Standard CLI / LCD / gRPC support included
   UI integrations to follow in later phases

9. Testnet Deployment & QA
   Deploy module on testnet for validation.
   Testing includes:
   Delegation / undelegation flows
   Reward accrual accuracy
   Validator status transitions: jailing, unjailing, tombstoning, and unbonded/unbonding status — and their effect on reward accrual, new delegation blocking, and forced unbonding
   Edge cases (partial unbonding, redelegation, redelegation in transit to tombstoned validator)
   Stress testing reward pool behavior including pool exhaustion
   EmergencyPause and resumption flow
   Publish results for validator and community review.

10. Documentation
   Module specification
   Validator guidance
   CLI / API usage examples
   Integration notes for wallets and dApps

Budget
Total: $34,000 USD (paid in LUNC at time of spend proposals)
This proposal focuses on technical implementation scope only.

Timeline
Week
Milestone
1–4
Module design and initial implementation
5–8
Core functionality development (delegation, rewards, unbonding)
9–10
Testnet deployment and QA
11
Address testnet and community review findings; fix issues; final code freeze
12
Documentation, validator coordination, governance proposal for activation (Phase 2)







Community Impact
Positive Effects:
Increased USTC utility
Reduced circulating supply pressure
More on-chain participation
Strengthened validator engagement
Neutral Effects:
No change to LUNC staking or validator power

Outcome
Upon completion, Terra Classic will:
Support native USTC staking on Layer 1
Provide a new utility layer for USTC holders
Maintain full consensus integrity
Establish a foundation for future revenue-driven rewards
