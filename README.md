# IX Sustainment OS

IX Sustainment OS is a **non-weapon, defense-adjacent sustainment operating system** for maintenance, readiness, parts bottlenecks, technical-data access, and auditable AI-assisted decision support.

It is designed to help sustainment organizations answer a simple but high-value question:

**What can we fix right now, what is blocked, why is it blocked, and what evidence supports the next action?**

---

## Why this exists

Sustainment organizations do not fail only because a part is missing.

They fail because the decision chain is fragmented across:
- maintenance records
- fault history
- technical manuals
- entitlement and data-rights boundaries
- parts availability
- approval chains
- inconsistent local workflows
- poor auditability of recommendations and overrides

IX Sustainment OS is intended to unify those fragments into one operational layer.

This repo is being built as a serious product concept for:
- depot maintenance environments
- field sustainment support workflows
- readiness and fleet-health teams
- program offices
- prime contractor sustainment teams
- controlled government or government-adjacent environments

---

## Product statement

**IX Sustainment OS** is a CUI-conscious sustainment platform that combines:

1. **Maintenance triage**
2. **Technical-data and entitlement checks**
3. **Parts and supply bottleneck visibility**
4. **Human-governed AI recommendations**
5. **Tamper-evident operational evidence**

The goal is not to replace maintainers, engineers, logisticians, or program staff.

The goal is to give them a shared operating picture with traceable, reviewable, decision-grade evidence.

---

## Core use cases

### 1) Maintenance triage
A maintainer, analyst, or operations lead needs to:
- ingest a fault or discrepancy
- classify severity and mission effect
- see similar prior cases
- determine if action is possible at current echelon
- identify blockers immediately

### 2) Technical-data access and entitlement
A user needs to know:
- which technical reference applies
- whether access is permitted
- whether a procedure is current
- whether a requested action crosses a data-rights or approval boundary

### 3) Parts and readiness bottlenecks
A sustainment planner needs to know:
- what parts are constraining readiness
- which shortages affect highest-priority assets
- which work orders are waiting on material, approval, tooling, or data
- where local fixes are being delayed by process, not physics

### 4) Human-governed AI assistance
A user needs machine support for:
- suggested triage paths
- likely fault families
- recommended next-step evidence collection
- probable bottleneck causes
- prioritization support

But every recommendation must remain:
- reviewable
- attributable
- overrideable
- auditable

### 5) Evidence for oversight and coordination
Leads and program stakeholders need a record of:
- who saw what
- which data informed a recommendation
- which user approved or rejected it
- what changed
- when it changed
- under which policy boundary it changed

---

## What this product is not

IX Sustainment OS is **not**:
- a weapon system
- a targeting platform
- a strike-planning tool
- a kill-chain optimization system
- a battlefield fires-control product
- an autonomous engagement system

It is also **not** intended to make unsupported claims such as:
- automatic authority to operate
- automatic CMMC compliance
- automatic NIST certification
- replacement of official logistics or maintenance systems of record

This project stays on the sustainment, readiness, workflow, auditability, and human-governed decision-support side.

---

## First-release product wedge

The first credible wedge is:

**“A sustainment triage and bottleneck operating layer for maintenance, parts, entitlement, and auditable AI recommendations.”**

That means Release 1 focuses on six modules:

### A. Case Intake
Structured fault and discrepancy intake with mission context, asset context, severity, and supporting evidence.

### B. Triage Board
A queue-based operational view for fault families, blockers, aging items, and recommended next actions.

### C. Technical Data Gateway
Controlled linking of procedures, references, revisions, access restrictions, and entitlement status.

### D. Parts Bottleneck Board
Visibility into work stoppage causes driven by supply, tooling, approval, or documentation constraints.

### E. Recommendation + Approval Layer
AI-assisted recommendations that remain human-reviewed, policy-bounded, and fully logged.

### F. Audit and Evidence Ledger
A durable record of actions, approvals, overrides, recommendation provenance, and state transitions.

---

## Primary users

### Maintainer / technician
Needs fast clarity on what to do next and what is blocking execution.

### Production controller / planner
Needs queue visibility, blocker trends, and work prioritization.

### Supply / logistics analyst
Needs shortage impact visibility and readiness consequence mapping.

### Sustainment engineer
Needs case history, recurring fault visibility, and procedure/evidence traceability.

### Program or operations lead
Needs readiness picture, delay cause concentration, and accountable decision history.

### Compliance / security reviewer
Needs role boundaries, policy enforcement, evidence trails, and reviewable AI behavior.

---

## Product principles

### 1. Human authority stays in the loop
Recommendations can assist, but accountable humans approve meaningful actions.

### 2. Evidence beats black boxes
Every important suggestion must trace back to observable inputs, policy context, and user action.

### 3. Workflow over dashboard theater
This product must reduce friction in real operational flow, not just look impressive in screenshots.

### 4. Policy-aware by default
Data exposure, recommendations, and workflow actions must respect access and policy boundaries.

### 5. Sustainment first
Every feature should improve one of these:
- maintenance velocity
- blocker identification
- readiness clarity
- evidence quality
- coordination speed

---

## High-level architecture

IX Sustainment OS is planned as a modular web platform with policy-bounded services.

### Planned layers

#### 1. Experience layer
Mission-critical operator UI for:
- queue triage
- case review
- approvals
- parts impact views
- evidence inspection

#### 2. Application layer
Business logic for:
- case lifecycle
- blocker classification
- workflow transitions
- approvals
- recommendation routing

#### 3. Policy and trust layer
Controls for:
- role-based access
- entitlement checks
- recommendation gating
- approval requirements
- action logging

#### 4. Data layer
Structured objects for:
- assets
- cases
- discrepancies
- procedures
- parts constraints
- approvals
- evidence events

#### 5. Integration layer
Connectors for:
- maintenance systems
- reference libraries
- supply feeds
- identity providers
- export/report pipelines

---

## Planned object model

The repo will be built around a small, serious domain model.

### Core entities
- **Asset**
- **Case**
- **FaultEvent**
- **ProcedureRef**
- **EntitlementRule**
- **PartConstraint**
- **Recommendation**
- **ApprovalDecision**
- **EvidenceEvent**
- **UserRole**
- **PolicyBundle**

### Core states
A case will generally move through states such as:
- `new`
- `triage`
- `awaiting-data`
- `awaiting-parts`
- `awaiting-approval`
- `actionable`
- `deferred`
- `resolved`
- `closed`

---

## AI boundary

AI in this platform is **assistive**, not sovereign.

AI may help with:
- classification support
- similarity matching
- next-step suggestion
- blocker explanation
- prioritization hints
- summarization of case history

AI may **not** silently:
- execute irreversible actions
- bypass approvals
- bypass access policy
- rewrite audit history
- promote itself to decision authority

The system must preserve:
- recommendation provenance
- confidence or uncertainty markers
- user override capability
- reviewable change history

---

## CUI and security posture

This project is being shaped for **CUI-conscious environments**, which means the design intent includes:
- least-privilege access boundaries
- strong auditability
- role-aware data exposure
- secure defaults
- reviewable workflow actions
- deployment patterns compatible with controlled environments

Important:
**This repo does not claim certification, accreditation, or compliance by declaration alone.**
Those outcomes depend on deployment, controls, environment, documentation, and assessment.

---

## Repo build plan

This repository will be completed in a staged manner across focused commits that establish:

1. product scope
2. license and commercial posture
3. repository structure
4. architecture documentation
5. operator workflows
6. domain schema
7. API surface
8. policy and audit model
9. backend scaffolding
10. UI scaffolding
11. demo data
12. compliance and deployment documentation

---

## Planned repository structure

```text
IX-Sustainment-OS/
├── README.md
├── LICENSE
├── COMMERCIAL_TERMS.md
├── .gitignore
├── Makefile
├── docs/
│   ├── architecture/
│   ├── product/
│   ├── workflows/
│   ├── security/
│   ├── compliance/
│   └── ux/
├── api/
│   └── openapi.yaml
├── schemas/
│   ├── case.schema.json
│   ├── recommendation.schema.json
│   ├── approval.schema.json
│   └── evidence-event.schema.json
├── cmd/
│   └── ix-sustainment-os/
├── internal/
│   ├── domain/
│   ├── policy/
│   ├── audit/
│   ├── entitlement/
│   ├── recommendation/
│   └── workflow/
├── web/
│   ├── app/
│   └── components/
├── demo/
│   ├── fixtures/
│   └── screenshots/
└── .github/
    └── workflows/
```

---

## Commercial direction

The commercial objective is straightforward:

- allow serious evaluation
- protect monetizable production value
- preserve leverage for pilots, support, and enterprise deployment

The planned licensing posture for this repo is:
- **Business Source License 1.1** for the public codebase
- **separate commercial terms** for production, pilot, support, and controlled deployment use

The actual license text and commercial-terms file will be added in follow-on commits.

---

## Success criteria

This repo is successful if a serious buyer can look at it and say:

- this solves a real sustainment workflow problem
- this is disciplined, not vague
- this respects policy and audit boundaries
- this could be piloted without pretending to be a weapon or autonomous command layer
- this has enough architecture, workflow clarity, and product seriousness to justify a conversation

---

## Current status

This is the founding commit.

It establishes:
- product identity
- operating boundary
- user problem framing
- first-release wedge
- architecture direction
- commercial posture
- repo shape

All implementation, policy files, schemas, and scaffolding follow from here.
