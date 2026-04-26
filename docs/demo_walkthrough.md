# Demo Walkthrough (60-90 sec)

Use this script for owner/investor/partner demos of the sandbox.

## 0) Setup (10 sec)

1. Ensure sandbox seed is loaded.
2. Open Backoffice UI.
3. Click **Load Production-like Sandbox**.

## 1) Happy Path (30-40 sec)

Scenario: **Buy 100 USDT**

1. Click **Load Scenario**.
2. Click **Run Full Flow**.
3. Narrate while watching cards:
   - Quote Summary: base/client rate, fee, TTL.
   - Order Summary: reserved/completed and version.
   - Timeline: quote -> reserve -> complete.
   - Why/Audit: source, rule, last action.

## 2) Error Path (20-25 sec)

Scenario: **Expired Quote** (or **Stale Rate**)

1. Click scenario chip.
2. Click **Run Reserve** (or **Run Quote** for stale).
3. Show:
   - status badges switch to reject/error state,
   - timeline marks failed step,
   - Why/Audit shows reason + last error.

## 3) Presentation Mode (5 sec)

1. Click **Hide Debug / Show Debug**.
2. Show only summary cards + timeline for clean business-facing view.

## 4) Close (5 sec)

Click **Reset Demo State** to prepare next run.
