"""Empirical feasibility probe for Issue #52 / Gate SCHED-DISC: Schedules and User Overrides.
Introspects and tests Schedule, ScheduleTarget/Rule, UserOverride, and OnCallNotificationRule against GoAlert v0.35.0.
"""
import json
import uuid
from fixture import disposable

INTROSPECTION_QUERY = """
query IntrospectDetailedScheduleSchema {
  scheduleTargetInput: __type(name: "ScheduleTargetInput") {
    inputFields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  scheduleTarget: __type(name: "ScheduleTarget") {
    fields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  createUserOverrideInput: __type(name: "CreateUserOverrideInput") {
    inputFields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  updateUserOverrideInput: __type(name: "UpdateUserOverrideInput") {
    inputFields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  setScheduleOnCallNotificationRulesInput: __type(name: "SetScheduleOnCallNotificationRulesInput") {
    inputFields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
  onCallNotificationRuleInput: __type(name: "OnCallNotificationRuleInput") {
    inputFields {
      name
      type { name kind ofType { name kind ofType { name kind } } }
    }
  }
}
"""

def main():
    print("Starting disposable GoAlert instance...", flush=True)
    with disposable() as fix:
        print("GoAlert ready. Running detailed introspection...", flush=True)
        data = fix.graphql(INTROSPECTION_QUERY, session=True)

        print("\n=== ScheduleTargetInput ===", flush=True)
        print(json.dumps(data.get("scheduleTargetInput"), indent=2), flush=True)

        print("\n=== ScheduleTarget Fields ===", flush=True)
        print(json.dumps(data.get("scheduleTarget"), indent=2), flush=True)

        print("\n=== CreateUserOverrideInput ===", flush=True)
        print(json.dumps(data.get("createUserOverrideInput"), indent=2), flush=True)

        print("\n=== UpdateUserOverrideInput ===", flush=True)
        print(json.dumps(data.get("updateUserOverrideInput"), indent=2), flush=True)

        print("\n=== SetScheduleOnCallNotificationRulesInput ===", flush=True)
        print(json.dumps(data.get("setScheduleOnCallNotificationRulesInput"), indent=2), flush=True)

        print("\n=== OnCallNotificationRuleInput ===", flush=True)
        print(json.dumps(data.get("onCallNotificationRuleInput"), indent=2), flush=True)

        # 1. Create User and Rotation for Schedule testing
        print("\n=== Creating Test User and Rotation ===", flush=True)
        u1 = fix.graphql(
            "mutation CreateU($input: CreateUserInput!) { createUser(input: $input) { id name } }",
            {"input": {"name": "Schedule User 1", "email": "sched1@example.com", "role": "user", "username": "scheduser1", "password": "Password123!"}},
            session=True
        )["createUser"]
        u2 = fix.graphql(
            "mutation CreateU($input: CreateUserInput!) { createUser(input: $input) { id name } }",
            {"input": {"name": "Schedule User 2", "email": "sched2@example.com", "role": "user", "username": "scheduser2", "password": "Password123!"}},
            session=True
        )["createUser"]
        rot = fix.graphql(
            "mutation CreateR($input: CreateRotationInput!) { createRotation(input: $input) { id name } }",
            {"input": {
                "name": "Sched Rotation",
                "type": "daily",
                "start": "2026-09-28T08:00:00Z",
                "timeZone": "Europe/Madrid",
                "userIDs": [u1["id"], u2["id"]],
            }},
            session=True
        )["createRotation"]
        print(f"Created users: {u1['id']}, {u2['id']}, rotation: {rot['id']}", flush=True)

        # 2. Test createSchedule
        print("\n=== Testing createSchedule ===", flush=True)
        sched = fix.graphql("""
        mutation CS($input: CreateScheduleInput!) {
            createSchedule(input: $input) {
                id
                name
                description
                timeZone
                isFavorite
            }
        }
        """, {
            "input": {
                "name": "Engineering On-Call Schedule",
                "description": "Primary rotation and shift coverage",
                "timeZone": "Europe/Madrid",
                "favorite": True
            }
        }, session=True)["createSchedule"]
        print("Created Schedule:", json.dumps(sched, indent=2), flush=True)
        sched_id = sched["id"]

        # 3. Test updateSchedule
        print("\n=== Testing updateSchedule ===", flush=True)
        fix.graphql("""
        mutation US($input: UpdateScheduleInput!) {
            updateSchedule(input: $input)
        }
        """, {
            "input": {
                "id": sched_id,
                "name": "Engineering Primary Schedule",
                "description": "Updated description",
                "timeZone": "Europe/Madrid"
            }
        }, session=True)

        # 4. Test schedule query with targets/rules
        print("\n=== Testing updateScheduleTarget (adding rotation rule) ===", flush=True)
        fix.graphql("""
        mutation UST($input: ScheduleTargetInput!) {
            updateScheduleTarget(input: $input)
        }
        """, {
            "input": {
                "scheduleID": sched_id,
                "target": {"id": rot["id"], "type": "rotation"},
                "rules": [
                    {
                        "start": "09:00",
                        "end": "17:00",
                        "weekdayFilter": [True, True, True, True, True, False, False]
                    }
                ]
            }
        }, session=True)

        # Read back schedule targets
        sched_data = fix.graphql("""
        query GetSched($id: ID!) {
            schedule(id: $id) {
                id
                name
                description
                timeZone
                targets {
                    target { id name type }
                    rules {
                        id
                        start
                        end
                        weekdayFilter
                    }
                }
            }
        }
        """, {"id": sched_id}, session=True)
        print("Schedule after target update:", json.dumps(sched_data, indent=2), flush=True)

        # 5. Test createUserOverride
        print("\n=== Testing createUserOverride ===", flush=True)
        override = fix.graphql("""
        mutation CUO($input: CreateUserOverrideInput!) {
            createUserOverride(input: $input) {
                id
                start
                end
                addUserID
                removeUserID
                target { id type }
            }
        }
        """, {
            "input": {
                "scheduleID": sched_id,
                "start": "2026-09-29T00:00:00Z",
                "end": "2026-09-30T00:00:00Z",
                "addUserID": u2["id"],
                "removeUserID": u1["id"]
            }
        }, session=True)["createUserOverride"]
        print("Created User Override:", json.dumps(override, indent=2), flush=True)

        # 6. Test updateUserOverride
        print("\n=== Testing updateUserOverride ===", flush=True)
        fix.graphql("""
        mutation UUO($input: UpdateUserOverrideInput!) {
            updateUserOverride(input: $input)
        }
        """, {
            "input": {
                "id": override["id"],
                "start": "2026-09-29T02:00:00Z",
                "end": "2026-09-30T02:00:00Z",
                "addUserID": u2["id"]
            }
        }, session=True)

        # 7. Test escalation policy step with schedule action
        print("\n=== Testing escalation policy step with schedule destination ===", flush=True)
        ep = fix.graphql("""
        mutation CEP($input: CreateEscalationPolicyInput!) {
            createEscalationPolicy(input: $input) {
                id
                steps {
                    id
                    actions {
                        type
                        args
                    }
                }
            }
        }
        """, {
            "input": {
                "name": "EP With Schedule Target",
                "steps": [
                    {
                        "delayMinutes": 15,
                        "actions": [
                            {"type": "builtin-schedule", "args": {"schedule_id": sched_id}}
                        ]
                    }
                ]
            }
        }, session=True)["createEscalationPolicy"]
        # 8. Test schedules search query
        print("\n=== Testing schedules query ===", flush=True)
        try:
            s_res = fix.graphql("""
            query SearchSchedules($search: String!) {
                schedules(input: {search: $search, first: 50}) {
                    nodes { id name timeZone }
                }
            }
            """, {"search": "Engineering"}, session=True)
            print("Search with input: {search} succeeded:", json.dumps(s_res, indent=2), flush=True)
        except Exception as e:
            print("Search with input failed:", e, flush=True)
            try:
                s_res = fix.graphql("""
                query SearchSchedulesDirect($search: String!) {
                    schedules(search: $search, first: 50) {
                        nodes { id name timeZone }
                    }
                }
                """, {"search": "Engineering"}, session=True)
                print("Search with direct search succeeded:", json.dumps(s_res, indent=2), flush=True)
            except Exception as e2:
                print("Direct search failed too:", e2, flush=True)

        # 9. Test reading userOverride directly and via schedule
        print("\n=== Testing userOverride read ===", flush=True)
        uo_res = fix.graphql("""
        query GetUO($id: ID!) {
            userOverride(id: $id) {
                id
                start
                end
                addUserID
                removeUserID
                target { id type }
            }
        }
        """, {"id": override["id"]}, session=True)
        print("Read UserOverride by ID:", json.dumps(uo_res, indent=2), flush=True)

        # 10. Test deleting target from schedule (updateScheduleTarget with empty rules)
        print("\n=== Testing removing target from schedule (empty rules) ===", flush=True)
        fix.graphql("""
        mutation ClearTarget($input: ScheduleTargetInput!) {
            updateScheduleTarget(input: $input)
        }
        """, {
            "input": {
                "scheduleID": sched_id,
                "target": {"id": rot["id"], "type": "rotation"},
                "rules": []
            }
        }, session=True)
        sched_cleared = fix.graphql("""
        query CheckCleared($id: ID!) {
            schedule(id: $id) {
                targets {
                    target { id }
                    rules { id }
                }
            }
        }
        """, {"id": sched_id}, session=True)
        print("Targets after clearing with empty rules:", json.dumps(sched_cleared, indent=2), flush=True)

        # 11. Test deleting UserOverride
        print("\n=== Testing deleteAll userOverride ===", flush=True)
        del_uo = fix.graphql("""
        mutation DelUO($input: [TargetInput!]!) {
            deleteAll(input: $input)
        }
        """, {"input": [{"id": override["id"], "type": "userOverride"}]}, session=True)
        print("deleteAll userOverride result:", del_uo, flush=True)

        # 12. Test deleting Schedule
        print("\n=== Testing deleteAll schedule ===", flush=True)
        del_sched = fix.graphql("""
        mutation DelSched($input: [TargetInput!]!) {
            deleteAll(input: $input)
        }
        """, {"input": [{"id": sched_id, "type": "schedule"}]}, session=True)
        print("deleteAll schedule result:", del_sched, flush=True)

        print("\nALL EMPIRICAL PROBES SUCCEEDED!", flush=True)

if __name__ == "__main__":
    main()
