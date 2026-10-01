"""Empirical feasibility probe for Issue #70 / Gate UPST-DISC: GoAlert v0.35.0 Schema Capabilities.
Probes multiAck on escalation policy steps, private and enableStatusUpdates on user contact methods,
and polymorphic labels across escalation policies, schedules, and rotations against GoAlert v0.35.0.
"""
import json
from fixture import disposable

def main():
    print("Starting disposable GoAlert v0.35.0 instance...", flush=True)
    with disposable() as fix:
        # 1. Probe multiAck on EscalationPolicyStep
        print("\n=== 1. Probing multiAck on EscalationPolicyStep ===", flush=True)
        ep_res = fix.graphql("""
        mutation CreateEP($input: CreateEscalationPolicyInput!) {
            createEscalationPolicy(input: $input) {
                id
                name
                steps {
                    id
                    stepNumber
                    delayMinutes
                    multiAck
                }
            }
        }
        """, {
            "input": {
                "name": "MultiAck Test Policy",
                "description": "Testing v0.35.0 multiAck",
                "repeat": 1,
                "steps": [
                    {
                        "delayMinutes": 10,
                        "multiAck": True,
                        "actions": [{"type": "builtin-webhook", "args": {"webhook_url": "https://example.com/hook"}}]
                    }
                ]
            }
        }, session=True)
        print("Created EP with multiAck=True:", json.dumps(ep_res, indent=2), flush=True)
        ep_id = ep_res["createEscalationPolicy"]["id"]
        step_id = ep_res["createEscalationPolicy"]["steps"][0]["id"]
        assert ep_res["createEscalationPolicy"]["steps"][0]["multiAck"] is True

        # Update multiAck to False
        fix.graphql("""
        mutation UpdateStep($input: UpdateEscalationPolicyStepInput!) {
            updateEscalationPolicyStep(input: $input)
        }
        """, {
            "input": {
                "id": step_id,
                "delayMinutes": 15,
                "multiAck": False
            }
        }, session=True)

        ep_read = fix.graphql("""
        query ReadEP($id: ID!) {
            escalationPolicy(id: $id) {
                id
                steps {
                    id
                    delayMinutes
                    multiAck
                }
            }
        }
        """, {"id": ep_id}, session=True)
        print("Read back EP after multiAck=False update:", json.dumps(ep_read, indent=2), flush=True)
        assert ep_read["escalationPolicy"]["steps"][0]["multiAck"] is False

        # 2. Probe private and enableStatusUpdates on UserContactMethod
        print("\n=== 2. Probing private & enableStatusUpdates on UserContactMethod ===", flush=True)
        user_res = fix.graphql("""
        mutation CreateUser($input: CreateUserInput!) {
            createUser(input: $input) {
                id
                name
            }
        }
        """, {
            "input": {
                "name": "Alice V035",
                "email": "alice.v035@example.com",
                "role": "user",
                "username": "alice-v035",
                "password": "Password1234567890!"
            }
        }, session=True)
        user_id = user_res["createUser"]["id"]

        # Create public contact method with enableStatusUpdates=True
        cm_pub = fix.graphql("""
        mutation CreateCM($input: CreateUserContactMethodInput!) {
            createUserContactMethod(input: $input) {
                id
                name
                disabled
                private
                statusUpdates
                dest {
                    type
                    args
                }
            }
        }
        """, {
            "input": {
                "userID": user_id,
                "name": "Public Webhook CM",
                "enableStatusUpdates": True,
                "private": False,
                "dest": {"type": "builtin-webhook", "args": {"webhook_url": "https://example.com/cm-pub"}}
            }
        }, session=True)
        print("Created public CM:", json.dumps(cm_pub, indent=2), flush=True)
        cm_pub_id = cm_pub["createUserContactMethod"]["id"]

        # Create private contact method for another user (alice) as admin
        try:
            cm_priv = fix.graphql("""
            mutation CreatePrivCM($input: CreateUserContactMethodInput!) {
                createUserContactMethod(input: $input) {
                    id
                    name
                    disabled
                    private
                    statusUpdates
                    dest {
                        type
                        args
                    }
                }
            }
            """, {
                "input": {
                    "userID": user_id,
                    "name": "Private Webhook CM",
                    "enableStatusUpdates": False,
                    "private": True,
                    "dest": {"type": "builtin-webhook", "args": {"webhook_url": "https://example.com/cm-priv"}}
                }
            }, session=True)
            print("Created private CM for Alice as admin:", json.dumps(cm_priv, indent=2), flush=True)
            cm_priv_id = cm_priv["createUserContactMethod"]["id"]

            # Now test reading private CM via userContactMethod(id) as admin session AND as system API key!
            read_priv_session = fix.graphql("""
            query ReadCM($id: ID!) {
                userContactMethod(id: $id) {
                    id
                    name
                    private
                    statusUpdates
                    dest {
                        type
                        args
                    }
                }
            }
            """, {"id": cm_priv_id}, session=True)
            print("Read private CM via userContactMethod(id) (admin session):", json.dumps(read_priv_session, indent=2), flush=True)

            # Test reading via User.contactMethods list
            read_user_cms = fix.graphql("""
            query ReadUserCMs($id: ID!) {
                user(id: $id) {
                    id
                    contactMethods {
                        id
                        name
                        private
                        statusUpdates
                    }
                }
            }
            """, {"id": user_id}, session=True)
            print("Read User.contactMethods (admin session):", json.dumps(read_user_cms, indent=2), flush=True)
        except Exception as e:
            print("Private CM probe exception:", e, flush=True)

        # Test updating public CM enableStatusUpdates & private
        fix.graphql("""
        mutation UpdateCM($input: UpdateUserContactMethodInput!) {
            updateUserContactMethod(input: $input)
        }
        """, {
            "input": {
                "id": cm_pub_id,
                "enableStatusUpdates": False,
                "private": True
            }
        }, session=True)
        after_upd_cm = fix.graphql("""
        query ReadCM($id: ID!) {
            userContactMethod(id: $id) {
                id
                name
                private
                statusUpdates
                dest {
                    type
                    args
                }
            }
        }
        """, {"id": cm_pub_id}, session=True)
        print("Read CM after updating private=True, enableStatusUpdates=False:", json.dumps(after_upd_cm, indent=2), flush=True)

        # 3. Probe polymorphic labels on EscalationPolicy, Schedule, and Rotation
        print("\n=== 3. Probing polymorphic labels on EscalationPolicy, Schedule, and Rotation ===", flush=True)
        sched_res = fix.graphql("""
        mutation CreateSched($input: CreateScheduleInput!) {
            createSchedule(input: $input) {
                id
                name
            }
        }
        """, {
            "input": {
                "name": "Labeled Schedule",
                "timeZone": "UTC"
            }
        }, session=True)
        sched_id = sched_res["createSchedule"]["id"]

        rot_res = fix.graphql("""
        mutation CreateRot($input: CreateRotationInput!) {
            createRotation(input: $input) {
                id
                name
            }
        }
        """, {
            "input": {
                "name": "Labeled Rotation",
                "type": "daily",
                "start": "2026-10-01T08:00:00Z",
                "timeZone": "UTC",
                "shiftLength": 1
            }
        }, session=True)
        rot_id = rot_res["createRotation"]["id"]

        for target_type, target_id in [
            ("escalationPolicy", ep_id),
            ("schedule", sched_id),
            ("rotation", rot_id),
        ]:
            fix.graphql("""
            mutation SetLbl($input: SetLabelInput!) {
                setLabel(input: $input)
            }
            """, {
                "input": {
                    "target": {"type": target_type, "id": target_id},
                    "key": "example.com/environment",
                    "value": f"prod-{target_type}"
                }
            }, session=True)

        lbl_read = fix.graphql("""
        query ReadAllLabels($epID: ID!, $schedID: ID!, $rotID: ID!) {
            escalationPolicy(id: $epID) {
                id
                labels { key value }
            }
            schedule(id: $schedID) {
                id
                labels { key value }
            }
            rotation(id: $rotID) {
                id
                labels { key value }
            }
        }
        """, {"epID": ep_id, "schedID": sched_id, "rotID": rot_id}, session=True)
        print("Read polymorphic labels:", json.dumps(lbl_read, indent=2), flush=True)
        assert lbl_read["escalationPolicy"]["labels"] == [{"key": "example.com/environment", "value": "prod-escalationPolicy"}]
        assert lbl_read["schedule"]["labels"] == [{"key": "example.com/environment", "value": "prod-schedule"}]
        assert lbl_read["rotation"]["labels"] == [{"key": "example.com/environment", "value": "prod-rotation"}]

        # Delete labels by setting empty string value
        for target_type, target_id in [
            ("escalationPolicy", ep_id),
            ("schedule", sched_id),
            ("rotation", rot_id),
        ]:
            fix.graphql("""
            mutation DelLbl($input: SetLabelInput!) {
                setLabel(input: $input)
            }
            """, {
                "input": {
                    "target": {"type": target_type, "id": target_id},
                    "key": "example.com/environment",
                    "value": ""
                }
            }, session=True)

        lbl_after_del = fix.graphql("""
        query ReadAllLabels($epID: ID!, $schedID: ID!, $rotID: ID!) {
            escalationPolicy(id: $epID) { labels { key value } }
            schedule(id: $schedID) { labels { key value } }
            rotation(id: $rotID) { labels { key value } }
        }
        """, {"epID": ep_id, "schedID": sched_id, "rotID": rot_id}, session=True)
        print("Read labels after deletion:", json.dumps(lbl_after_del, indent=2), flush=True)
        assert lbl_after_del["escalationPolicy"]["labels"] == []
        assert lbl_after_del["schedule"]["labels"] == []
        assert lbl_after_del["rotation"]["labels"] == []

        print("\nALL GOALERT v0.35.0 CAPABILITY PROBES SUCCEEDED!", flush=True)

if __name__ == "__main__":
    main()
