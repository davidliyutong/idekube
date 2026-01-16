package idekube.rbac

import future.keywords.if
import future.keywords.in

# Default deny
default allow := false

# Allow if there's a direct policy match
allow if {
    policy := data.policies[_]
    policy.subject == input.subject
    policy.object == input.object
    policy.action == input.action
}

# Allow if subject has a role that grants permission
allow if {
    # Get roles for the subject
    role := data.role_bindings[input.subject][_]
    
    # Check if role has permission
    policy := data.policies[_]
    policy.subject == role
    matches_object(policy.object, input.object)
    matches_action(policy.action, input.action)
}

# Allow if wildcard policy matches
allow if {
    role := data.role_bindings[input.subject][_]
    policy := data.policies[_]
    policy.subject == role
    policy.object == "*"
    matches_action(policy.action, input.action)
}

# Allow super admin full access
allow if {
    role := data.role_bindings[input.subject][_]
    role == "role:super_admin"
}

# Helper: check if object pattern matches
matches_object(pattern, object) if {
    pattern == object
}

matches_object(pattern, object) if {
    pattern == "*"
}

# Support for resource ownership patterns like "workspace:own"
matches_object(pattern, object) if {
    contains(pattern, ":own")
    split_pattern := split(pattern, ":")
    split_object := split(object, ":")
    split_pattern[0] == split_object[0]
}

# Support for keyMatch-style patterns
matches_object(pattern, object) if {
    glob.match(pattern, ["/"], object)
}

# Helper: check if action pattern matches
matches_action(pattern, action) if {
    pattern == action
}

matches_action(pattern, action) if {
    pattern == "*"
}

# Get all roles for a subject
roles_for_subject[role] {
    role := data.role_bindings[input.subject][_]
}

# Get all permissions for a role
permissions_for_role[permission] {
    role := input.role
    policy := data.policies[_]
    policy.subject == role
    permission := {
        "object": policy.object,
        "action": policy.action
    }
}
