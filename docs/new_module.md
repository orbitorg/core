
Terra Classic USTC Staking System -Technical Implementation Proposal
Following Signal Proposal 12219 (USTC Staking System), which passed, this proposal defines the technical scope, architecture, and implementation plan for introducing a USTC staking system on Layer 1.

Executive Summary
Following the successful signal from Proposal 12219, this proposal outlines the design and implementation of a native USTC staking system on Terra Classic Layer 1.
The objective is to introduce staking-like utility for USTC holders without modifying the core validator consensus model. This will be achieved by implementing a new dedicated module (x/ustcstaking), enabling users to delegate USTC, earn rewards, and participate in the ecosystem , while preserving LUNC as the sole consensus staking asset.
This approach ensures:
Zero impact on consensus security
Clean modular design aligned with Cosmos SDK standards
Scalable foundation for future enhancements (oracle tuning, fee-based rewards, etc.)

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

2. Core Functionality
   Delegation System
   Users delegate uusd (USTC) to validators
   Delegations tracked independently from LUNC staking
   Validator serves as accounting endpoint only

Unbonding Logic
Configurable unbonding period
Optional unbond with penalty (parameter[eg 7days])
Reward Distribution
Rewards distributed proportionally based on USTC stake
Epoch-based distribution
Independent reward pool
Reward Pool
Funded via governance-controlled mechanisms
3. Parameters (Governance Controlled)
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
   oracle_enabled (disabled in Phase 1)
   All parameters adjustable via governance.

4. Message Types
   MsgDelegateUSTC
   MsgUndelegateUSTC
   MsgBeginRedelegateUSTC
   MsgWithdrawUSTCRewards
   MsgFundUSTCRewardPool
   MsgUpdateUSTCParams
   MsgEmergencyPause

5. Reward Model (Phase 1)
   Fixed Reward Model (Initial Implementation):
   Rewards emitted from a pre-funded pool
   APR set conservatively via governance
   Distribution proportional to stake
   This ensures:
   Simplicity
   Predictability
   Controlled rollout

6. Oracle Mechanism (Future Phase)
   A separate lightweight oracle system may be introduced in later phases to:
   Suggest APR adjustments
   Monitor staking participation
   Provide validator-submitted signals
   This will:
   NOT interfere with Terra’s existing price oracle
   Be bounded by governance-defined limits

7. Integration
   Validators
   Act as delegation endpoints
   May receive commission on USTC rewards
   No impact on consensus power
   Wallets & dApps
   Not required for Phase 1
   Standard CLI / LCD / gRPC support included
   UI integrations to follow in later phases

8. Testnet Deployment & QA
   Deploy module on testnet for validation.
   Testing includes:
   Delegation / undelegation flows
   Reward accrual accuracy
   Edge cases (partial unbonding, redelegation)
   Stress testing reward pool behavior
   Publish results for validator and community review.

9. Documentation
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
Documentation and community review
12
Governance proposal for activation (Phase 2)







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

Conclusion
This proposal delivers a safe, modular, and scalable implementation of USTC staking, aligned with the direction approved by the community in Proposal 12219.
It prioritizes:
Simplicity
Security
Incremental rollout
We invite the Terra Classic community to support this technical implementation and participate in testing and validation.
Vegas & Orbit Labs