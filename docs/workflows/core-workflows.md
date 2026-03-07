# IX Sustainment OS — Core Workflows

## Purpose

This document defines the **operational workflows** that IX Sustainment OS must support in its first credible release.

It exists to make these things unambiguous:

- who does what
- when a case changes state
- what can block progress
- what requires approval
- where AI may assist
- what evidence must be recorded
- how the UI should behave under real operational pressure

This file is not a user manual. It is a **product-and-systems workflow specification**.

---

## First-release workflow objective

The first release must prove one serious workflow claim:

**IX Sustainment OS can take a sustainment issue from intake to defensible next action while preserving blocker visibility, policy boundaries, and auditability.**

That means the system must support a complete path across:

1. case intake  
2. triage  
3. blocker identification  
4. technical-data check  
5. parts and supply check  
6. recommendation review  
7. approval or override  
8. actionability determination  
9. resolution tracking  
10. evidence preservation

---

## Primary operator roles

The first-release workflows assume these primary user roles.

### 1. Maintainer / technician
Primary concerns:
- what is wrong
- what should be checked next
- whether action can proceed
- what is blocking execution
- which procedure applies

### 2. Production controller / planner
Primary concerns:
- queue state
- aging items
- blockers by category
- work prioritization
- cases waiting on parts or approval

### 3. Supply / logistics analyst
Primary concerns:
- readiness effect of shortages
- repeated part constraints
- alternate sourcing implications
- which cases are materially blocked

### 4. Sustainment engineer / analyst
Primary concerns:
- recurring fault families
- evidence completeness
- similar prior cases
- procedure relevance
- technical-data applicability

### 5. Lead / supervisor / approver
Primary concerns:
- approval items
- override decisions
- queue risk
- critical asset impact
- decision accountability

### 6. Security / policy reviewer
Primary concerns:
- access boundaries
- recommendation controls
- action policy enforcement
- evidence completeness
- restricted-data handling behavior

---

## First-release workflow map

```text
Case Intake
   ↓
Initial Triage
   ↓
Blocker Identification
   ↓
┌───────────────────────────────────────────────────────┐
│ Parallel checks                                       │
│  - technical-data / procedure check                   │
│  - parts / supply constraint check                    │
│  - recommendation generation (if allowed)             │
└───────────────────────────────────────────────────────┘
   ↓
Disposition Decision
   ├── actionable
   ├── awaiting-data
   ├── awaiting-parts
   ├── awaiting-approval
   ├── deferred
   └── resolved / closed

This is the first public wedge. The system does not need every future workflow at once. It needs this one to feel disciplined and real.

Authoritative case states

The first release should use a controlled case state model.

new

Meaning:

case has been created

minimal normalization may have occurred

no meaningful triage outcome has been accepted yet

Entry conditions:

user submits case

integration creates case

imported discrepancy enters the queue

Exit conditions:

case enters triage

triage

Meaning:

case is under active initial assessment

severity, mission effect, and likely blocker lane are being established

Entry conditions:

new case accepted into workflow

reopened case returned for reassessment

Exit conditions:

actionable

awaiting-data

awaiting-parts

awaiting-approval

deferred

awaiting-data

Meaning:

insufficient evidence, reference clarity, or diagnostic input exists to proceed safely

Examples:

missing diagnostic evidence

unclear fault signature

conflicting observations

relevant procedure cannot yet be confirmed

Exit conditions:

triage

actionable

deferred

awaiting-parts

Meaning:

progress is currently blocked by material or supply constraint

Examples:

required part unavailable

substitute unresolved

supply ETA unknown

allocation pending

Exit conditions:

triage

actionable

deferred

awaiting-approval

Meaning:

progress depends on human review or policy-gated authorization

Examples:

recommendation requires lead approval

override requires accountable sign-off

restricted procedure link requires approval

priority escalation requires authorization

Exit conditions:

actionable

triage

deferred

actionable

Meaning:

the case has a valid next action path with no unresolved blocker preventing controlled execution

Examples:

procedure confirmed

entitlement satisfied

required evidence sufficient

parts available or not required

approvals complete

Exit conditions:

resolved

triage

deferred

deferred

Meaning:

work is intentionally not proceeding now

Examples:

operational priority is lower

dependencies outside current control

accepted wait state

scheduled later action

Exit conditions:

triage

closed

resolved

Meaning:

the case has been worked to a satisfactory operational conclusion

Examples:

corrective action completed

blocker cleared and no further action needed

validated disposition recorded

Exit conditions:

closed

triage if reopened

closed

Meaning:

the case is complete from workflow perspective

no additional action is expected unless reopened

Exit conditions:

triage if reopened under controlled action

State transition rules

The system should avoid casual or ambiguous transitions.

Valid state transition matrix

new               -> triage
triage            -> awaiting-data
triage            -> awaiting-parts
triage            -> awaiting-approval
triage            -> actionable
triage            -> deferred
awaiting-data     -> triage
awaiting-data     -> actionable
awaiting-data     -> deferred
awaiting-parts    -> triage
awaiting-parts    -> actionable
awaiting-parts    -> deferred
awaiting-approval -> actionable
awaiting-approval -> triage
awaiting-approval -> deferred
actionable        -> resolved
actionable        -> triage
actionable        -> deferred
deferred          -> triage
deferred          -> closed
resolved          -> closed
resolved          -> triage
closed            -> triage

Invalid examples

These should be blocked or at least strongly guarded:

new -> resolved

new -> closed

awaiting-parts -> closed

awaiting-data -> resolved

awaiting-approval -> resolved

A case should not appear to have skipped essential operational reasoning.

Blocker taxonomy

The platform must represent blockers as a first-class operational concept.

First-release blocker categories
1. data

The case is blocked because evidence or diagnostic context is insufficient.

Examples:

incomplete fault details

missing prior-event context

no supporting attachment

unresolved ambiguity

2. procedure

The case is blocked because the applicable procedure or reference is unknown, outdated, or not yet confirmed.

Examples:

no matching procedure found

procedure revision conflict

procedure relevance uncertain

3. entitlement

The case is blocked because access or use of required technical data is restricted for the current user or lane.

Examples:

user lacks access

restricted document boundary

approval needed before exposing reference

4. parts

The case is blocked by a material or supply issue.

Examples:

part unavailable

no alternate identified

backorder

distribution hold

5. tooling

The case is blocked by lack of required equipment or support capability.

Examples:

tooling unavailable

equipment calibration issue

test rig access delayed

6. approval

The case is blocked by a required human decision or policy gate.

Examples:

override pending

escalation pending

restricted action pending review

7. capacity

The case is blocked by queue or staffing pressure rather than technical impossibility.

Examples:

work-center overload

review backlog

constrained planner bandwidth

8. policy

The case is blocked because the requested action violates policy constraints.

Examples:

prohibited state change

unauthorized role action

unapproved workflow branch

The UI must allow more than one blocker on a case, but one blocker should be markable as primary.

Workflow 1 — Case intake
Objective

Capture a sustainment issue in a structured way without slowing operators down.

Trigger

A user or integration submits a new issue.

Inputs

Required minimums:

asset identifier

issue title or summary

reported condition

severity estimate

reporter identity or source

time of observation

Useful optional inputs:

subsystem area

mission effect

attachment or image

suspected fault family

reference to prior event

urgency note

operating context

System behavior

The platform should:

create case record

assign a unique case identifier

normalize timestamps

attach asset context if known

record reporter/source

set state to new

emit case-created evidence event

route to triage queue

UX requirements

The intake form must be:

fast

structured

scan-friendly

safe for partial context

explicit about required vs optional fields

The system should support:

manual entry

prefilled integration entry

draft-safe submission flow if applicable

Audit events

case.created

asset.linked

attachment.added

source.recorded

Workflow 2 — Initial triage
Objective

Turn a raw intake into a defensible preliminary operational posture.

Trigger

A case in new is opened by a triage-capable user or triage automation lane.

Operator questions

how severe is this

is the mission effect meaningful

does this resemble a known issue

what is the most likely blocker category

can action proceed now

System behavior

The platform should:

move case to triage

expose key case summary

show relevant asset metadata

show linked prior similar cases if available

allow blocker tagging

allow severity refinement

allow priority assignment

optionally generate bounded recommendations

AI assistance allowed

The platform may assist by:

summarizing case context

retrieving similar cases

suggesting likely blocker categories

highlighting missing evidence

proposing next diagnostic data to collect

AI assistance not allowed

The platform may not:

finalize state without review

assign authoritative approval outcome

silently change priority or severity

suppress contradictory evidence

Exit outcomes

From triage, the user should be able to send the case to:

awaiting-data

awaiting-parts

awaiting-approval

actionable

deferred

Audit events

case.state_changed

severity.updated

priority.updated

blocker.tagged

recommendation.generated

recommendation.viewed

Workflow 3 — Technical-data and procedure check
Objective

Determine whether the correct procedure or reference exists, applies, and is authorized for use.

Trigger

A triage user, engineer, or recommendation requests a relevant procedure.

Key questions

what technical reference is relevant

is it current

does it apply to this asset or configuration

is the user entitled to access it

does viewing or using it require approval

System behavior

The platform should:

search or link relevant procedure refs

display revision metadata

display applicability metadata

check entitlement rules

indicate one of:

accessible

restricted

unknown

outdated

conflict

record the result

Decision outcomes

Possible results:

procedure confirmed and accessible

likely procedure identified but restricted

conflicting references found

no adequate reference found

reference requires approval before use

State implications

Common outcomes:

proceed toward actionable

move to awaiting-approval

move to awaiting-data

Audit events

procedure.search_performed

procedure.linked

procedure.access_denied

procedure.applicability_confirmed

entitlement.check_performed

Workflow 4 — Parts and supply bottleneck check
Objective

Determine whether material availability is blocking execution and whether readiness impact is concentrated.

Trigger

The case appears to need a part, supply action, or material confirmation.

Key questions

is a required part known

is it available

is there an alternate path

which other cases are blocked by the same shortage

what is the readiness consequence

System behavior

The platform should:

attach one or more part constraints to the case

show availability status

show estimated delay if known

show concentration across cases

distinguish material block from non-material block

allow planners or analysts to annotate impact

Decision outcomes

Possible results:

no material blocker exists

part is available, case may proceed

part is unavailable, move to awaiting-parts

supply ambiguity exists, more information required

alternate path exists but requires approval

Audit events

part_constraint.added

part_constraint.updated

supply_status.checked

readiness_impact.assessed

Workflow 5 — Recommendation review
Objective

Make assistive AI useful without allowing it to become invisible authority.

Trigger

A recommendation is generated or surfaced for operator review.

Recommendation types

First-release recommendation classes:

likely fault family

next evidence to collect

likely blocker cause

relevant procedure suggestion

queue-priority hint

similar-case retrieval suggestion

Required recommendation fields

Every recommendation shown to a user should preserve:

recommendation ID

recommendation type

rationale summary

inputs used

generation timestamp

confidence or uncertainty marker

approval requirement flag

current status

Review actions

A reviewer should be able to:

accept

reject

override

defer

request escalation

Rules

accepted recommendations do not silently rewrite history

rejected recommendations remain inspectable

overrides require accountable actor identity

certain recommendation classes may require explicit approval before creating state changes

Audit events

recommendation.generated

recommendation.reviewed

recommendation.accepted

recommendation.rejected

recommendation.overridden

Workflow 6 — Approval flow
Objective

Preserve accountability for sensitive or policy-gated decisions.

Trigger examples

recommendation requires lead approval

restricted procedure link requested

state transition crosses policy boundary

override requested

priority escalation requested

Approval item fields

The approval object should include:

approval ID

related object type

related object ID

requested action

requester

approver role requirement

reason

evidence references

due time if applicable

status

Approval decisions

Allowed dispositions:

approved

rejected

returned for more information

State behavior

Common outcomes:

awaiting-approval -> actionable

awaiting-approval -> triage

awaiting-approval -> deferred

Rules

approvals must be attributable

silent approvals are forbidden

approval decision should capture reason

expired or stale approval items should surface visibly

Audit events

approval.requested

approval.viewed

approval.approved

approval.rejected

approval.returned_for_info

Workflow 7 — Actionability decision
Objective

Declare when the case has a valid next action path.

Conditions for actionable

A case should usually be marked actionable only when:

evidence is sufficient for current decision scope

required procedure path is known

entitlement boundary is satisfied

required parts are available or not needed

required approvals are complete

no blocking policy conflict remains

UX expectations

When a case becomes actionable, the interface should make clear:

why it is actionable

what blockers were cleared

who made the determination

what the next operator step is

Audit events

case.actionable_marked

blocker.cleared

approval.boundary_satisfied

Workflow 8 — Resolution and closure
Objective

Preserve a trustworthy end-state for the case.

Resolution conditions

A case may be marked resolved when:

required corrective action or disposition is complete

final notes are entered

key supporting evidence is attached or referenced

operator or lead has recorded final assessment

Closure conditions

A case may be marked closed when:

no further action is expected

the workflow record is complete enough for later review

resolution state is stable

any post-resolution checks have been satisfied if required

Reopen rules

A closed case may be reopened only by:

allowed roles

explicit actor identity

recorded reason

Audit events

case.resolved

case.closed

case.reopened

Queue and board behavior

The first-release operator experience should include a queue or board view with these essential slices.

Core queue slices

all open cases

critical or high-severity cases

awaiting-data

awaiting-parts

awaiting-approval

actionable now

stale or aging cases

cases with repeated blocker categories

Required visible columns or attributes

At minimum, the operator board should show:

case ID

asset

title / summary

severity

current state

primary blocker

age

priority

approval flag

recommendation flag

mission effect indicator if applicable

UX principle

The board should answer:

what needs attention first

what is blocked and why

what can move now

what is waiting on human decision

Aging and escalation rules

The system should treat time as an operational factor.

First-release aging indicators

The platform should visibly mark:

newly created but untouched cases

long-running cases in triage

approval items aging beyond expected window

prolonged parts-blocked cases

stale deferred cases

Escalation examples

case in triage too long

approval request not reviewed

high-severity case still awaiting data

repeated shortage affecting high-priority assets

Escalation does not need to auto-resolve anything. It needs to create visibility and accountability.

Example end-to-end scenario
Scenario

A new issue is reported against an important asset. The issue appears correctable, but progress may be blocked by both reference access and material constraints.

Flow

maintainer creates case with summary, asset, severity, and note

system records case.created

case enters new

triage user opens case; system moves it to triage

recommendation layer suggests likely blocker is procedure plus part constraint

technical-data gateway finds a relevant procedure, but it is restricted for the current user

system records entitlement-check result

parts check identifies likely required part, currently unavailable

triage user marks primary blocker as parts and secondary blocker as entitlement

case moves to awaiting-parts

supervisor requests approval to expose restricted procedure summary to allowed role

approval is granted for limited review path

later, part becomes available

blockers clear

case moves to actionable

work completes

case moves to resolved

final review occurs

case moves to closed

That case now has a defensible operational trail, not just a final status label.

Mandatory evidence trail expectations

The first release should preserve evidence for the following workflow moments:

case creation

state change

severity change

priority change

blocker add/remove

procedure link or denial

entitlement check

recommendation generation

recommendation review

approval request and decision

part constraint link/update

actionable determination

resolution

closure

reopen

This is not optional decoration. It is part of the product value.

UX constraints derived from the workflows

The workflows imply several UX rules.

Rule 1 — state must always be obvious

Users must not hunt for the current case status.

Rule 2 — blockers must be visible without opening every detail

Primary blockers should be legible at queue level.

Rule 3 — approvals must be hard to miss

Sensitive pending actions should not disappear into a generic notification area.

Rule 4 — AI output must look assistive, not authoritative

Recommendations must be visually distinct from confirmed facts and human decisions.

Rule 5 — action history must be inspectable quickly

The user should be able to see what changed, who changed it, and why.

Workflow anti-patterns to avoid
1. Fake actionability

The system marks a case actionable even though a real blocker still exists.

2. Recommendation authority creep

The recommendation appears to decide the outcome by default.

3. Hidden entitlement failure

The user cannot tell whether the needed procedure is missing or merely restricted.

4. Dashboard without workflow consequence

The system visualizes bottlenecks but offers no path to act.

5. Ambiguous closure

The case is closed without enough evidence to defend that closure later.

These are serious product failure modes and should be treated as design constraints.

Definition of workflow success

The workflow design is successful if a serious reviewer can say:

the state model is disciplined

the blocker model is useful

the approval flow is accountable

the AI boundary is controlled

the queue behavior matches real operator needs

the product could plausibly help a sustainment team move faster without losing trust

Status

This file establishes the baseline workflow model for the repo.

Follow-on commits should translate this workflow model into:

domain schemas

API routes

audit events

policy checks

queue UI structure

demo fixtures
