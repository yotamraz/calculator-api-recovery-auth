#!/usr/bin/env python3
"""
BE Testing - API Contract Validation Tests

Generated pytest script to validate the API spec against the running application.
Each endpoint is tested as a parameterized test case using pytest.

This script supports two modes:
1. SRC Validation: Tests endpoints and captures responses (no expected_response)
2. DST Contract Validation: Tests endpoints and validates responses match expected (has expected_response)

Generated at: 2026-02-17T17:14:00.557697+00:00
Project: calculator-api-recovery-auth
Milestone: 68
"""

import base64
import json
import os
import sys
import textwrap
import time
from typing import Any

import pytest
import requests

# =============================================================================
# Test Configuration (embedded from spec validation)
# =============================================================================

def resolve_env_placeholders(obj: Any) -> Any:
    """Recursively resolve ${VAR_NAME} environment variable placeholders in test data.

    The spec refinement agent replaces detected secrets with ${VAR_NAME} placeholders.
    The actual secret values are passed as environment variables by the runner.
    """
    if isinstance(obj, str):
        return os.path.expandvars(obj)
    if isinstance(obj, dict):
        return {k: resolve_env_placeholders(v) for k, v in obj.items()}
    if isinstance(obj, list):
        return [resolve_env_placeholders(item) for item in obj]
    return obj


# Parse JSON at runtime, then resolve any ${VAR_NAME} env var placeholders
# that the agent may have substituted for detected secrets.
TEST_CASES: list[dict[str, Any]] = resolve_env_placeholders(
    json.loads(r'''[
    {
        "name": "health_check_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/health",
        "method": "GET",
        "description": "Verify health endpoint returns ok status and version 0.1.0",
        "request_data": {
            "path": {},
            "query": {},
            "body": null
        },
        "expected_status": 200,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "register_user_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/auth/register",
        "method": "POST",
        "description": "Register a new user with valid username and password",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "username": "testuser_reg_hp",
                "password": "${TEST_USER_PASSWORD}"
            }
        },
        "expected_status": 201,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "register_user_duplicate_username",
        "category": "INVALID_INPUT",
        "endpoint": "/auth/register",
        "method": "POST",
        "description": "Attempt to register with a username that already exists, expect 400",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "username": "testuser_dup_check",
                "password": "${TEST_USER_PASSWORD}"
            }
        },
        "expected_status": 400,
        "setup": {
            "endpoint": "/auth/register",
            "method": "POST",
            "body": {
                "username": "testuser_dup_check",
                "password": "${TEST_USER_PASSWORD}"
            }
        },
        "cleanup": null
    },
    {
        "name": "register_user_missing_password",
        "category": "MISSING_REQUIRED",
        "endpoint": "/auth/register",
        "method": "POST",
        "description": "Attempt to register with missing password field, expect 422 validation error",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "username": "testuser_no_pw"
            }
        },
        "expected_status": 422,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "login_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/auth/token",
        "method": "POST",
        "description": "Login with valid credentials using OAuth2 form data and receive a JWT token",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "username": "testuser_login_hp",
                "password": "${TEST_USER_PASSWORD}"
            },
            "content_type": "application/x-www-form-urlencoded"
        },
        "expected_status": 200,
        "setup": {
            "endpoint": "/auth/register",
            "method": "POST",
            "body": {
                "username": "testuser_login_hp",
                "password": "${TEST_USER_PASSWORD}"
            }
        },
        "cleanup": null
    },
    {
        "name": "login_invalid_credentials",
        "category": "INVALID_INPUT",
        "endpoint": "/auth/token",
        "method": "POST",
        "description": "Login with incorrect password, expect 401 Unauthorized",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "username": "testuser_login_bad",
                "password": "wrong_password_xyz"
            },
            "content_type": "application/x-www-form-urlencoded"
        },
        "expected_status": 401,
        "setup": {
            "endpoint": "/auth/register",
            "method": "POST",
            "body": {
                "username": "testuser_login_bad",
                "password": "${TEST_USER_PASSWORD}"
            }
        },
        "cleanup": null
    },
    {
        "name": "login_nonexistent_user",
        "category": "INVALID_INPUT",
        "endpoint": "/auth/token",
        "method": "POST",
        "description": "Login with a username that does not exist, expect 401",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "username": "nonexistent_user_xyz",
                "password": "anypassword"
            },
            "content_type": "application/x-www-form-urlencoded"
        },
        "expected_status": 401,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "add_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/add",
        "method": "POST",
        "description": "Add two numbers with valid authentication",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "a": 10.5,
                "b": 5.3
            },
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 200,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "add_no_auth",
        "category": "INVALID_INPUT",
        "endpoint": "/add",
        "method": "POST",
        "description": "Attempt to add without authentication token, expect 401",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "a": 1.0,
                "b": 2.0
            }
        },
        "expected_status": 401,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "subtract_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/subtract",
        "method": "POST",
        "description": "Subtract two numbers with valid authentication",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "a": 20.0,
                "b": 7.5
            },
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 200,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "multiply_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/multiply",
        "method": "POST",
        "description": "Multiply two numbers with valid authentication",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "a": 4.0,
                "b": 3.0
            },
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 200,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "divide_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/divide",
        "method": "POST",
        "description": "Divide two numbers with valid authentication",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "a": 15.0,
                "b": 3.0
            },
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 200,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "divide_by_zero",
        "category": "BOUNDARY",
        "endpoint": "/divide",
        "method": "POST",
        "description": "Attempt division by zero, expect 400 with error message",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "a": 10.0,
                "b": 0
            },
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 400,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "divide_no_auth",
        "category": "INVALID_INPUT",
        "endpoint": "/divide",
        "method": "POST",
        "description": "Attempt to divide without authentication token, expect 401",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "a": 10.0,
                "b": 2.0
            }
        },
        "expected_status": 401,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "create_calculation_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/calculations",
        "method": "POST",
        "description": "Create a new calculation record with valid operation and authentication",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "operation": "add",
                "a": 8.0,
                "b": 4.0
            },
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 201,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "create_calculation_unknown_operation",
        "category": "INVALID_INPUT",
        "endpoint": "/calculations",
        "method": "POST",
        "description": "Attempt to create a calculation with an unknown operation, expect 400",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "operation": "modulo",
                "a": 10.0,
                "b": 3.0
            },
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 400,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "create_calculation_div_by_zero",
        "category": "BOUNDARY",
        "endpoint": "/calculations",
        "method": "POST",
        "description": "Attempt to create a division calculation with b=0, expect 400",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "operation": "div",
                "a": 10.0,
                "b": 0
            },
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 400,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "create_calculation_no_auth",
        "category": "INVALID_INPUT",
        "endpoint": "/calculations",
        "method": "POST",
        "description": "Attempt to create a calculation without auth, expect 401",
        "request_data": {
            "path": {},
            "query": {},
            "body": {
                "operation": "add",
                "a": 1.0,
                "b": 2.0
            }
        },
        "expected_status": 401,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "list_calculations_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/calculations",
        "method": "GET",
        "description": "List all stored calculations with valid authentication, expect 200 with array",
        "request_data": {
            "path": {},
            "query": {},
            "body": null,
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 200,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "list_calculations_no_auth",
        "category": "INVALID_INPUT",
        "endpoint": "/calculations",
        "method": "GET",
        "description": "Attempt to list calculations without authentication, expect 401",
        "request_data": {
            "path": {},
            "query": {},
            "body": null
        },
        "expected_status": 401,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "get_calculation_by_id_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/calculations/{calculation_id}",
        "method": "GET",
        "description": "Create a calculation, then retrieve it by ID",
        "request_data": {
            "path": {
                "calculation_id": "$setup_id"
            },
            "query": {},
            "body": null,
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 200,
        "setup": {
            "endpoint": "/calculations",
            "method": "POST",
            "body": {
                "operation": "mul",
                "a": 6.0,
                "b": 7.0
            },
            "auth": "bearer $fresh_access_token",
            "extract_id_from": "id"
        },
        "cleanup": {
            "endpoint": "/calculations/{calculation_id}",
            "method": "DELETE",
            "path": {
                "calculation_id": "$setup_id"
            },
            "auth": "bearer $fresh_access_token"
        }
    },
    {
        "name": "get_calculation_not_found",
        "category": "NOT_FOUND",
        "endpoint": "/calculations/{calculation_id}",
        "method": "GET",
        "description": "Attempt to retrieve a calculation that does not exist, expect 404",
        "request_data": {
            "path": {
                "calculation_id": 999999
            },
            "query": {},
            "body": null,
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 404,
        "setup": null,
        "cleanup": null
    },
    {
        "name": "delete_calculation_happy_path",
        "category": "HAPPY_PATH",
        "endpoint": "/calculations/{calculation_id}",
        "method": "DELETE",
        "description": "Create a calculation, then delete it by ID",
        "request_data": {
            "path": {
                "calculation_id": "$setup_id"
            },
            "query": {},
            "body": null,
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 204,
        "setup": {
            "endpoint": "/calculations",
            "method": "POST",
            "body": {
                "operation": "sub",
                "a": 100.0,
                "b": 25.0
            },
            "auth": "bearer $fresh_access_token",
            "extract_id_from": "id"
        },
        "cleanup": null
    },
    {
        "name": "delete_calculation_not_found",
        "category": "NOT_FOUND",
        "endpoint": "/calculations/{calculation_id}",
        "method": "DELETE",
        "description": "Attempt to delete a calculation that does not exist, expect 404",
        "request_data": {
            "path": {
                "calculation_id": 999999
            },
            "query": {},
            "body": null,
            "auth": "bearer $fresh_access_token"
        },
        "expected_status": 404,
        "setup": null,
        "cleanup": null
    }
]''')
)

# Static auth loaded from environment variables (secrets are never embedded in script)
STATIC_AUTH_HEADERS: dict[str, str] = json.loads(os.environ.get("AUTH_HEADERS", "{}"))
STATIC_AUTH_COOKIES: dict[str, str] = json.loads(os.environ.get("AUTH_COOKIES", "{}"))

# Base URL for API requests (from app discovery, includes host:port)
# Expand any environment variables (e.g., $PORT, ${HOST}) in the URL
BASE_URL = os.path.expandvars("http://localhost:8000")
HEALTH_CHECK_ENDPOINT = os.path.expandvars("health")
# Request timeout in seconds
REQUEST_TIMEOUT = 30
# Full health check URL (same as healthcheck_command uses)
HEALTH_CHECK_URL = f"{BASE_URL.rstrip('/')}/{HEALTH_CHECK_ENDPOINT.lstrip('/')}"
# Response validation mode: when True, validates response body against expected_response
# This is automatically detected based on presence of expected_response in test cases
VALIDATE_RESPONSE_BODY = any(
    tc.get("expected_response") is not None or tc.get("actual_response") is not None
    for tc in TEST_CASES
)

# =============================================================================
# Dynamic Authentication
# =============================================================================

# Cache for authentication result (computed once, reused for all tests)
_auth_cache: dict[str, Any] | None = None


def authenticate() -> dict[str, Any]:
    """
    Perform authentication and return headers/cookies for API requests.

    This function is called once before tests run. It should return a dict with:
    - "headers": dict[str, str] - Headers to add to each request (e.g., Authorization)
    - "cookies": dict[str, str] - Cookies to add to each request

    The auth function body is passed via the AUTH_FUNCTION environment variable
    (base64-encoded Python code). This avoids embedding secrets in the script.
    """
    _auth_code = os.environ.get("AUTH_FUNCTION", "")
    if _auth_code:
        _decoded = base64.b64decode(_auth_code).decode("utf-8")
        # Build a callable from the auth function body and execute it
        _ns: dict[str, Any] = {
            "os": os, "json": json, "requests": requests, "time": time,
            "BASE_URL": BASE_URL, "HEALTH_CHECK_URL": HEALTH_CHECK_URL,
            "REQUEST_TIMEOUT": REQUEST_TIMEOUT, "Any": Any,
        }
        exec(
            compile("def _fn():\n" + textwrap.indent(_decoded, "    "), "<auth>", "exec"),
            _ns,
        )
        return _ns["_fn"]()
    # Fallback: return static auth headers/cookies from environment
    return {"headers": dict(STATIC_AUTH_HEADERS), "cookies": dict(STATIC_AUTH_COOKIES)}


def get_auth() -> dict[str, Any]:
    """Get cached authentication info, computing it if needed."""
    global _auth_cache
    if _auth_cache is None:
        try:
            _auth_cache = authenticate()
            print(f"Authentication successful, got headers: {list(_auth_cache.get('headers', {}).keys())}")
        except Exception as e:
            print(f"Authentication failed: {e}, falling back to static auth")
            _auth_cache = {"headers": dict(STATIC_AUTH_HEADERS), "cookies": dict(STATIC_AUTH_COOKIES)}
    return _auth_cache


# =============================================================================
# Response Contract Validation Utilities
# =============================================================================

# Fields that are expected to differ between SRC and DST (dynamic values)
# These fields are checked for presence and type, but values are not compared
DYNAMIC_FIELDS = {
    "id", "uuid", "created_at", "updated_at", "timestamp", "modified_at",
    "created_date", "updated_date", "last_modified", "date_created", "date_updated",
    "_id", "createdAt", "updatedAt", "modifiedAt", "token", "access_token",
    "refresh_token", "session_id", "request_id", "trace_id", "correlation_id",
}



def get_type_name(value: Any) -> str:
    """Get a human-readable type name for a value."""
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "boolean"
    if isinstance(value, int):
        return "integer"
    if isinstance(value, float):
        return "number"
    if isinstance(value, str):
        return "string"
    if isinstance(value, list):
        return "array"
    if isinstance(value, dict):
        return "object"
    return type(value).__name__


def is_same_type(actual: Any, expected: Any) -> bool:
    """Check if two values have compatible types."""
    # None matches None
    if actual is None and expected is None:
        return True
    if actual is None or expected is None:
        return False

    # Bool is special - don't mix with int
    if isinstance(actual, bool) or isinstance(expected, bool):
        return isinstance(actual, bool) and isinstance(expected, bool)

    # Numbers are interchangeable (int/float)
    if isinstance(actual, (int, float)) and isinstance(expected, (int, float)):
        return True

    # Same type check
    return type(actual) == type(expected)


def validate_contract(
    actual: Any,
    expected: Any,
    path: str = "",
) -> tuple[bool, list[str]]:
    """
    Validate that actual response matches the expected response.

    This performs value comparison with special handling for dynamic fields:
    1. Structure: All expected fields must exist in actual
    2. Types: Values must have compatible types
    3. Values: Non-dynamic fields must have matching values
    4. Dynamic fields: Only checked for presence and type (values can differ)

    Dynamic fields (like id, created_at, token) are expected to differ between
    SRC and DST environments, so only their type is validated.

    Args:
        actual: The actual response from DST
        expected: The expected response (from SRC's actual_response)
        path: Current path in the response (for error messages)

    Returns:
        tuple: (is_valid, list of violations)
    """
    violations: list[str] = []

    # No expected = no validation needed
    if expected is None:
        return True, []

    # Actual is None but expected is not
    if actual is None:
        violations.append(f"{path or 'root'}: expected {get_type_name(expected)} but got null")
        return False, violations

    # Type validation
    if not is_same_type(actual, expected):
        violations.append(
            f"{path or 'root'}: type mismatch - expected {get_type_name(expected)}, "
            f"got {get_type_name(actual)}"
        )
        return False, violations

    # Object/dict validation - check structure and values
    if isinstance(expected, dict):
        for key, exp_val in expected.items():
            key_path = f"{path}.{key}" if path else key
            is_dynamic = key.lower() in {f.lower() for f in DYNAMIC_FIELDS}

            if key not in actual:
                violations.append(f"{key_path}: missing required field")
                continue

            act_val = actual[key]

            if is_dynamic:
                # For dynamic fields, only validate type (values will differ)
                if not is_same_type(act_val, exp_val):
                    violations.append(
                        f"{key_path}: type mismatch - expected {get_type_name(exp_val)}, "
                        f"got {get_type_name(act_val)}"
                    )
            else:
                # For non-dynamic fields, recursively validate (includes value comparison)
                valid, sub_violations = validate_contract(act_val, exp_val, key_path)
                violations.extend(sub_violations)

        return len(violations) == 0, violations

    # Array validation - check length and validate each item
    if isinstance(expected, list):
        # Length validation
        if len(actual) != len(expected):
            violations.append(
                f"{path or 'root'}: array length mismatch - expected {len(expected)} items, "
                f"got {len(actual)} items"
            )
            # Continue to validate what we can

        # Validate items that exist in both
        for i, (act_item, exp_item) in enumerate(zip(actual, expected)):
            item_path = f"{path}[{i}]" if path else f"[{i}]"
            valid, sub_violations = validate_contract(act_item, exp_item, item_path)
            violations.extend(sub_violations)

        return len(violations) == 0, violations

    # String validation - direct value comparison
    if isinstance(expected, str):
        if actual != expected:
            violations.append(
                f"{path or 'root'}: value mismatch - expected \"{expected}\", "
                f"got \"{actual}\""
            )
            return False, violations
        return True, []

    # Number validation - direct value comparison
    if isinstance(expected, (int, float)):
        if actual != expected:
            violations.append(
                f"{path or 'root'}: value mismatch - expected {expected}, got {actual}"
            )
            return False, violations
        return True, []

    # Boolean validation - direct value comparison
    if isinstance(expected, bool):
        if actual != expected:
            violations.append(
                f"{path or 'root'}: value mismatch - expected {expected}, got {actual}"
            )
            return False, violations
        return True, []

    return True, []


def format_response_diff(differences: list[str], max_items: int = 10) -> str:
    """Format response differences for error message."""
    if not differences:
        return "No differences"

    output = []
    for i, diff in enumerate(differences[:max_items]):
        output.append(f"  - {diff}")

    if len(differences) > max_items:
        output.append(f"  ... and {len(differences) - max_items} more differences")

    return "\n".join(output)


# =============================================================================
# Test Results Collection
# =============================================================================

test_results: list[dict[str, Any]] = []


def record_result(
    name: str,
    endpoint: str,
    method: str,
    expected_status: int,
    actual_status: int,
    passed: bool,
    duration_ms: float,
    category: str | None = None,
    description: str | None = None,
    error: str | None = None,
    response_body: str | None = None,
    response_match: bool | None = None,
    response_diff: list[str] | None = None,
) -> None:
    """Record a test result for final output."""
    result: dict[str, Any] = {
        "name": name,
        "endpoint": endpoint,
        "method": method,
        "expected_status": expected_status,
        "actual_status": actual_status,
        "passed": passed,
        "duration_ms": duration_ms,
        "category": category,
        "description": description,
    }
    if error:
        result["error"] = error

    # Track response validation results (for DST contract testing)
    if response_match is not None:
        result["response_match"] = response_match

    # Always capture response for passed tests (for validation), truncate for failed
    if response_body:
        if passed:
            # Try to parse as JSON for structured response
            try:
                result["response"] = json.loads(response_body)
            except json.JSONDecodeError:
                result["response_body"] = response_body
        else:
            result["response_body"] = response_body

    test_results.append(result)


# =============================================================================
# Response Store (cross-test-case value sharing)
# =============================================================================

# In-memory store for values extracted from responses and shared across test cases.
# Test cases with a "store" field extract values from their response and save them here.
# Later test cases reference stored values via "$stored.KEY" placeholders.
_response_store: dict[str, Any] = {}


def extract_by_json_path(data: Any, json_path: str) -> Any:
    """Extract a value from nested data using a dot-separated JSON path.

    Supports dict key access and integer list indexing.
    E.g. "data.users.0.id" -> data["data"]["users"][0]["id"]
    """
    current = data
    for key in json_path.split("."):
        if current is None:
            return None
        if isinstance(current, dict):
            current = current.get(key)
        elif isinstance(current, list):
            try:
                current = current[int(key)]
            except (ValueError, IndexError):
                return None
        else:
            return None
    return current


def store_response_values(test_case: dict[str, Any], response_json: Any) -> None:
    """Extract values from a response and save them in the response store.

    The test case's "store" field maps placeholder names to dot-separated JSON
    paths in the response body. After a successful request, the extracted values
    become available to subsequent test cases via "$stored.<name>" placeholders.
    """
    store_config = test_case.get("store")
    if not store_config or not isinstance(store_config, dict):
        return
    if not isinstance(response_json, (dict, list)):
        return

    for placeholder_name, json_path in store_config.items():
        if not isinstance(json_path, str):
            continue
        value = extract_by_json_path(response_json, json_path)
        if value is not None:
            _response_store[placeholder_name] = value
            print(f"  Stored: ${placeholder_name} = {str(value)[:80]}...")
        else:
            print(f"  Warning: store path '{json_path}' resolved to None for '{placeholder_name}'")


def update_auth_from_response(
    test_case: dict[str, Any],
    response_json: Any,
    response: "requests.Response | None" = None,
) -> None:
    """Update the global auth cache from a test response.

    Called after a successful test that has ``store_auth`` configuration.
    This lets login/register tests propagate authentication to all subsequent tests.

    ``store_auth`` format::

        {
            "headers": {
                "Authorization": "Bearer {access_token}"   # {key} resolved from response store
            },
            "cookies": {
                "session_id": "session_id"                 # JSON path into response body
            },
            "from_cookies": true                            # also persist Set-Cookie headers
        }

    If the test case has no ``store_auth`` but its category is "SETUP" or "AUTH",
    we auto-detect common auth tokens in the response and update the cache.
    """
    global _auth_cache
    store_auth = test_case.get("store_auth")
    category = (test_case.get("category") or "").upper()

    if not store_auth and category not in ("SETUP", "AUTH"):
        return

    if _auth_cache is None:
        _auth_cache = {"headers": {}, "cookies": {}}

    # --- Explicit store_auth configuration ---
    if store_auth and isinstance(store_auth, dict):
        # Update headers (supports {placeholder} interpolation from _response_store)
        for header_name, template in store_auth.get("headers", {}).items():
            resolved = str(template)
            # Resolve {key} placeholders from response store
            for key, value in _response_store.items():
                resolved = resolved.replace(f"{{{key}}}", str(value))
            _auth_cache.setdefault("headers", {})[header_name] = resolved

        # Update cookies from response body paths
        for cookie_name, json_path in store_auth.get("cookies", {}).items():
            if isinstance(response_json, (dict, list)):
                value = extract_by_json_path(response_json, json_path)
                if value is not None:
                    _auth_cache.setdefault("cookies", {})[cookie_name] = str(value)

        # Persist Set-Cookie headers from the HTTP response
        if store_auth.get("from_cookies") and response is not None:
            for cookie_name, cookie_value in response.cookies.items():
                _auth_cache.setdefault("cookies", {})[cookie_name] = cookie_value

        print(
            f"  Auth updated (explicit): headers={list(_auth_cache.get('headers', {}).keys())}, "
            f"cookies={list(_auth_cache.get('cookies', {}).keys())}"
        )
        return

    # --- Auto-detect auth from SETUP/AUTH category tests ---
    if isinstance(response_json, dict):
        # Common token field names
        token_fields = {
            "access_token", "token", "accessToken", "jwt",
            "id_token", "idToken", "auth_token", "authToken",
        }
        for field in token_fields:
            value = response_json.get(field)
            if value and isinstance(value, str):
                # Determine prefix (Bearer for most tokens)
                prefix = "Bearer " if field != "jwt" else ""
                _auth_cache.setdefault("headers", {})["Authorization"] = f"{prefix}{value}"
                print(f"  Auth auto-detected: Authorization header set from '{field}'")
                break

        # Also persist cookies from the response
        if response is not None:
            for cookie_name, cookie_value in response.cookies.items():
                _auth_cache.setdefault("cookies", {})[cookie_name] = cookie_value
                print(f"  Auth auto-detected: cookie '{cookie_name}' persisted")


def resolve_stored_placeholders(obj: Any) -> Any:
    """Replace $stored.KEY placeholders with values from the response store.

    Handles three cases:
    1. Exact match: value is "$stored.key" -> replaced with stored value (preserves type)
    2. Embedded match: value is "Bearer $stored.token" -> string interpolation
    3. Recursive: dicts and lists are traversed recursively
    """
    if not _response_store:
        return obj

    if isinstance(obj, str):
        # Exact match - preserves original type (e.g. int, dict) instead of stringifying
        if obj.startswith("$stored."):
            key = obj[len("$stored."):]
            if key in _response_store:
                return _response_store[key]
        # Embedded string interpolation (handles "Bearer $stored.token" and partial matches)
        if "$stored." in obj:
            result = obj
            for key, value in _response_store.items():
                result = result.replace(f"$stored.{key}", str(value))
            return result
        return obj

    if isinstance(obj, dict):
        return {k: resolve_stored_placeholders(v) for k, v in obj.items()}

    if isinstance(obj, list):
        return [resolve_stored_placeholders(item) for item in obj]

    return obj


# =============================================================================
# Helper Functions
# =============================================================================


def build_request_url(endpoint: str, path_params: dict[str, Any], query_params: dict[str, Any]) -> str:
    """Build the full request URL with path and query parameters."""
    # Substitute path parameters
    for param_name, param_value in path_params.items():
        endpoint = endpoint.replace(f"{{{param_name}}}", str(param_value))

    # Add query parameters
    if query_params:
        query_string = "&".join(f"{k}={v}" for k, v in query_params.items())
        endpoint = f"{endpoint}?{query_string}"

    return f"{BASE_URL.rstrip('/')}{endpoint}"


def build_headers(request_headers: dict[str, str], skip_auth: bool = False) -> dict[str, str]:
    """Build request headers, optionally including auth."""
    if skip_auth:
        # Return only the explicit request headers, no auth
        return dict(request_headers)
    auth = get_auth()
    headers = dict(auth.get("headers", {}))
    headers.update(request_headers)
    return headers


def build_cookies(skip_auth: bool = False) -> dict[str, str]:
    """Build request cookies, optionally including auth."""
    if skip_auth:
        # Return empty cookies, no auth
        return {}
    auth = get_auth()
    return dict(auth.get("cookies", {}))


def make_request(
    method: str,
    endpoint: str,
    path_params: dict[str, Any],
    query_params: dict[str, Any],
    body: Any,
    request_headers: dict[str, str] | None = None,
    skip_auth: bool = False,
) -> requests.Response:
    """Make an HTTP request and return the response."""
    url = build_request_url(endpoint, path_params, query_params)
    headers = build_headers(request_headers or {}, skip_auth=skip_auth)
    cookies = build_cookies(skip_auth=skip_auth)

    method = method.upper()
    if method == "GET":
        return requests.get(url, headers=headers, cookies=cookies, timeout=REQUEST_TIMEOUT)
    elif method == "POST":
        return requests.post(url, headers=headers, cookies=cookies, json=body, timeout=REQUEST_TIMEOUT)
    elif method == "PUT":
        return requests.put(url, headers=headers, cookies=cookies, json=body, timeout=REQUEST_TIMEOUT)
    elif method == "PATCH":
        return requests.patch(url, headers=headers, cookies=cookies, json=body, timeout=REQUEST_TIMEOUT)
    elif method == "DELETE":
        return requests.delete(url, headers=headers, cookies=cookies, timeout=REQUEST_TIMEOUT)
    else:
        return requests.request(method, url, headers=headers, cookies=cookies, json=body, timeout=REQUEST_TIMEOUT)


def _extract_id_from_json(data: Any, extract_path: str) -> str | None:
    """Extract an ID from JSON data using a dot-separated path.

    Handles nested paths like "data.id" and returns the value as a string.
    Returns None if the path cannot be resolved.
    """
    try:
        current = data
        for key in extract_path.split("."):
            if isinstance(current, dict):
                current = current[key]
            elif isinstance(current, list):
                current = current[int(key)]
            else:
                return None
        return str(current)
    except (KeyError, TypeError, IndexError, ValueError):
        return None


def run_setup(setup_config: dict[str, Any]) -> str | None:
    """Run a setup request and return the extracted ID.

    Handles 'already exists' responses (409/422) gracefully by:
    1. Trying to extract the ID from the error response body
    2. Falling back to a GET request to find the existing resource
    """
    endpoint = setup_config.get("endpoint", "/")
    method = setup_config.get("method", "POST")
    body = resolve_stored_placeholders(setup_config.get("body"))
    extract_id_from = setup_config.get("extract_id_from", "id")

    resp = make_request(method, endpoint, {}, {}, body)

    # Happy path — resource created successfully
    if resp.status_code < 400:
        extracted = _extract_id_from_json(resp.json() if resp.text else {}, extract_id_from)
        if extracted is not None:
            return extracted
        print(f"Setup: Created resource but could not extract ID via path '{extract_id_from}'")
        return None

    # Handle "already exists" (409 Conflict, 422 Unprocessable Entity)
    if resp.status_code in (409, 422):
        print(f"Setup: Resource may already exist ({resp.status_code}). Attempting recovery...")

        # Strategy 1: Try to extract ID from the error response body
        # Many APIs include the existing resource info in the conflict response
        try:
            error_data = resp.json()
            extracted = _extract_id_from_json(error_data, extract_id_from)
            if extracted is not None:
                print(f"  Recovered existing resource ID from error response: {extracted}")
                return extracted
        except (json.JSONDecodeError, TypeError):
            pass

        # Strategy 2: Try a GET request on the same endpoint to find existing resource
        print(f"  Trying GET {endpoint} to find existing resource...")
        try:
            get_resp = make_request("GET", endpoint, {}, {}, None)
            if get_resp.status_code < 400 and get_resp.text:
                get_data = get_resp.json()
                # Handle list responses — take the last created item
                if isinstance(get_data, list) and get_data:
                    get_data = get_data[-1]
                extracted = _extract_id_from_json(get_data, extract_id_from)
                if extracted is not None:
                    print(f"  Recovered existing resource ID via GET: {extracted}")
                    return extracted
        except (json.JSONDecodeError, TypeError, requests.RequestException) as e:
            print(f"  GET fallback failed: {e}")

        print(f"Setup: Could not recover from 'already exists'. Response: {resp.text[:500]}")
        return None

    # Other error codes — not recoverable
    print(f"Setup failed: {resp.status_code} - {resp.text[:500]}")
    return None


def run_cleanup(cleanup_config: dict[str, Any], setup_id: str | None) -> None:
    """Run a cleanup request (best effort, errors are logged but don't fail)."""
    if not cleanup_config:
        return

    endpoint = cleanup_config.get("endpoint", "/")
    method = cleanup_config.get("method", "DELETE")
    path_params = dict(cleanup_config.get("path", {}))
    body = cleanup_config.get("body")

    # Replace $setup_id placeholder
    if setup_id:
        for key, val in path_params.items():
            if val == "$setup_id":
                path_params[key] = setup_id

    try:
        resp = make_request(method, endpoint, path_params, {}, body)
        if resp.status_code >= 400:
            print(f"Cleanup warning: {resp.status_code} - {resp.text}")
    except Exception as e:
        print(f"Cleanup warning: {e}")


# =============================================================================
# Pytest Fixtures
# =============================================================================

# Cache for fresh tokens (obtained from actual login)
_fresh_tokens_cache: dict[str, Any] | None = None


def get_fresh_tokens() -> dict[str, Any]:
    """
    Get fresh tokens by performing actual login.

    This is used for tests that need dynamically generated tokens (e.g., refresh token tests).
    Returns a dict with access_token, refresh_token, and other login response fields.
    """
    global _fresh_tokens_cache
    if _fresh_tokens_cache is not None:
        return _fresh_tokens_cache

    # Try to get tokens from the authenticate function
    try:
        auth_result = authenticate()
        # If authenticate() returns tokens directly, use them
        if "access_token" in auth_result or "refresh_token" in auth_result:
            _fresh_tokens_cache = auth_result
            print(f"Fresh tokens obtained from authenticate()")
            return _fresh_tokens_cache

        # Otherwise, we need to perform login ourselves
        # Look for login endpoint in test cases
        login_cases = [tc for tc in TEST_CASES if "login" in tc.get("name", "").lower() and tc.get("expected_status") == 200]
        if login_cases:
            login_case = login_cases[0]
            endpoint = login_case.get("endpoint", "/api/v1/auth/login")
            body = login_case.get("request_data", {}).get("body", {})

            url = f"{BASE_URL.rstrip('/')}{endpoint}"
            resp = requests.post(url, json=body, timeout=REQUEST_TIMEOUT)

            if resp.status_code == 200:
                _fresh_tokens_cache = resp.json()
                print(f"Fresh tokens obtained from login endpoint: {endpoint}")
                return _fresh_tokens_cache
            else:
                print(f"Login failed with status {resp.status_code}: {resp.text}")

        # Fallback: empty tokens
        _fresh_tokens_cache = {}
        return _fresh_tokens_cache

    except Exception as e:
        print(f"Failed to get fresh tokens: {e}")
        _fresh_tokens_cache = {}
        return _fresh_tokens_cache


def substitute_token_placeholders(body: Any) -> Any:
    """
    Replace token placeholders in request body with fresh tokens.

    Supported placeholders:
    - $fresh_refresh_token: Replaced with actual refresh_token from login
    - $fresh_access_token: Replaced with actual access_token from login
    """
    if body is None:
        return None

    if isinstance(body, str):
        if "$fresh_refresh_token" in body or "$fresh_access_token" in body:
            tokens = get_fresh_tokens()
            body = body.replace("$fresh_refresh_token", tokens.get("refresh_token", ""))
            body = body.replace("$fresh_access_token", tokens.get("access_token", ""))
        return body

    if isinstance(body, dict):
        result = {}
        for key, value in body.items():
            if isinstance(value, str):
                if value == "$fresh_refresh_token":
                    tokens = get_fresh_tokens()
                    result[key] = tokens.get("refresh_token", "")
                elif value == "$fresh_access_token":
                    tokens = get_fresh_tokens()
                    result[key] = tokens.get("access_token", "")
                else:
                    result[key] = substitute_token_placeholders(value)
            else:
                result[key] = substitute_token_placeholders(value)
        return result

    if isinstance(body, list):
        return [substitute_token_placeholders(item) for item in body]

    return body


@pytest.fixture(scope="session", autouse=True)
def wait_for_app_health() -> None:
    """Wait for the application to be healthy before running tests."""
    # HEALTH_CHECK_URL is the full URL (e.g., http://localhost:8080/health)
    # This matches the healthcheck_command used by the remote runner
    print(f"\nWaiting for app to be healthy at {HEALTH_CHECK_URL}...")

    max_attempts = 60
    interval = 2
    for attempt in range(max_attempts):
        try:
            resp = requests.get(HEALTH_CHECK_URL, timeout=5)
            if resp.status_code < 400:
                print(f"App is healthy after {attempt + 1} attempts")
                return
        except Exception as e:
            if attempt % 10 == 0:
                print(f"Health check attempt {attempt + 1}/{max_attempts}: {e}")
        time.sleep(interval)

    pytest.fail(f"Application failed health check at {HEALTH_CHECK_URL} after {max_attempts} attempts")


# =============================================================================
# Test Cases
# =============================================================================


def get_test_ids() -> list[str]:
    """Generate test IDs for parametrization."""
    return [tc.get("name", f"test_{i}") for i, tc in enumerate(TEST_CASES)]


@pytest.mark.parametrize("test_case", TEST_CASES, ids=get_test_ids())
def test_api_endpoint(test_case: dict[str, Any]) -> None:
    """Test a single API endpoint based on test case configuration."""
    # Extract test case info first (these are safe operations)
    name = test_case.get("name", "unnamed")
    endpoint = test_case.get("endpoint", "/")
    method = test_case.get("method", "GET").upper()
    expected_status = test_case.get("expected_status", 200)
    request_data = test_case.get("request_data", {})
    category = test_case.get("category")
    description = test_case.get("description")
    setup_config = test_case.get("setup")
    cleanup_config = test_case.get("cleanup")
    skip_auth = test_case.get("skip_auth", False)

    # Expected response for DST contract validation (from SRC validation)
    # The field is stored as 'actual_response' from SRC validation
    # Also check 'expected_response' for backwards compatibility
    expected_response = test_case.get("actual_response") or test_case.get("expected_response")

    # Initialize setup_id outside try so it's available in finally for cleanup
    setup_id: str | None = None

    # Wrap ENTIRE test execution in try/except to capture any failure in JSON results
    # This ensures exceptions during setup/execution are always recorded
    try:
        # Run setup if configured (creates resource and extracts ID)
        if setup_config:
            setup_id = run_setup(setup_config)
            # Check if setup is required (default True) - handle both dict and non-dict setup_config
            setup_required = True
            if isinstance(setup_config, dict):
                setup_required = setup_config.get("required", True)
            if setup_id is None and setup_required:
                # Record the failure BEFORE calling pytest.fail so it appears in JSON results
                record_result(
                    name=name,
                    endpoint=endpoint,
                    method=method,
                    expected_status=expected_status,
                    actual_status=0,  # No request was made
                    passed=False,
                    duration_ms=0,
                    category=category,
                    description=description,
                    error=f"Setup failed - cannot create required resource for test '{name}'",
                )
                # FAIL instead of skip - all tests must run, agent can exclude tests later
                pytest.fail(f"Setup failed for test '{name}' - cannot create required resource")

        # Inner try for the actual test execution
        # Build request components
        path_params = dict(request_data.get("path", {}))
        query_params = dict(request_data.get("query", {}))
        req_headers = request_data.get("headers", {})
        body = request_data.get("body")

        # Replace $setup_id placeholder in path params
        if setup_id:
            for key, val in path_params.items():
                if val == "$setup_id":
                    path_params[key] = setup_id

        # Replace token placeholders in body (for tests needing fresh tokens)
        body = substitute_token_placeholders(body)

        # Resolve $stored.* placeholders from previous test responses
        path_params = resolve_stored_placeholders(path_params)
        query_params = resolve_stored_placeholders(query_params)
        req_headers = resolve_stored_placeholders(req_headers)
        body = resolve_stored_placeholders(body)
        endpoint = resolve_stored_placeholders(endpoint)

        # Execute request
        start_time = time.time()
        try:
            resp = make_request(method, endpoint, path_params, query_params, body, req_headers, skip_auth)
            duration_ms = (time.time() - start_time) * 1000
            actual_status = resp.status_code
            response_body = resp.text

            # Store response values for cross-test-case sharing (before any assertions)
            if status_passed := (actual_status == expected_status):
                try:
                    resp_json = json.loads(response_body)
                    store_response_values(test_case, resp_json)
                    # Propagate auth from login/register/setup tests to subsequent tests
                    update_auth_from_response(test_case, resp_json, resp)
                except (json.JSONDecodeError, TypeError):
                    pass  # Non-JSON responses can't be stored

            # Check status code first
            error_msg = None if status_passed else f"Expected status {expected_status}, got {actual_status}"

            # Check response contract if actual_response is provided (DST contract validation)
            # This validates structure, types, and field presence - NOT exact values
            response_match: bool | None = None
            response_diff: list[str] | None = None

            if expected_response is not None and status_passed:
                # Parse actual response as JSON for contract validation
                try:
                    actual_json = json.loads(response_body)
                    response_match, response_diff = validate_contract(
                        actual_json,
                        expected_response,
                    )
                    if not response_match:
                        error_msg = (
                            f"Response contract violation:\n{format_response_diff(response_diff)}"
                        )
                except json.JSONDecodeError:
                    response_match = False
                    response_diff = ["Failed to parse response as JSON"]
                    error_msg = "Response is not valid JSON but contract validation is required"

            # Overall pass: status must match, and if expected_response exists, it must match too
            passed = status_passed and (response_match is None or response_match)

            record_result(
                name=name,
                endpoint=endpoint,
                method=method,
                expected_status=expected_status,
                actual_status=actual_status,
                passed=passed,
                duration_ms=duration_ms,
                category=category,
                description=description,
                error=error_msg,
                response_body=response_body,
                response_match=response_match,
                response_diff=response_diff,
            )

            # Use pytest assertion for proper test reporting
            if not status_passed:
                pytest.fail(
                    f"Test '{name}': Expected status {expected_status}, got {actual_status}. "
                    f"Response: {resp.text if resp.text else 'empty'}"
                )

            if response_match is False:
                pytest.fail(
                    f"Test '{name}': Response contract violation (DST validation).\n"
                    f"The response structure/types don't match the expected contract.\n"
                    f"Violations:\n{format_response_diff(response_diff or [])}"
                )

        except requests.RequestException as e:
            duration_ms = (time.time() - start_time) * 1000
            record_result(
                name=name,
                endpoint=endpoint,
                method=method,
                expected_status=expected_status,
                actual_status=0,
                passed=False,
                duration_ms=duration_ms,
                category=category,
                description=description,
                error=str(e),
            )
            pytest.fail(f"Test '{name}': Request failed with error: {e}")

    except Exception as e:
        # Catch ANY exception (including setup errors, type errors, etc.)
        # Record it as a failure so it appears in JSON results
        record_result(
            name=name,
            endpoint=endpoint,
            method=method,
            expected_status=expected_status,
            actual_status=0,
            passed=False,
            duration_ms=0,
            category=category,
            description=description,
            error=f"Test error: {type(e).__name__}: {e}",
        )
        raise  # Re-raise for pytest to report

    finally:
        # Always run cleanup (even if test fails)
        if cleanup_config:
            run_cleanup(cleanup_config, setup_id)


# =============================================================================
# Test Results Output
# =============================================================================


@pytest.fixture(scope="session", autouse=True)
def output_test_results(request: pytest.FixtureRequest) -> Any:
    """Output test results in JSON format after all tests complete."""
    yield  # Wait for all tests to complete

    # Calculate final results
    passed_count = sum(1 for r in test_results if r["passed"])
    failed_count = len([r for r in test_results if not r["passed"]])
    total_count = len(test_results)
    all_passed = failed_count == 0 and total_count > 0

    failures = [r for r in test_results if not r["passed"]]

    # Count response validation results (for DST contract testing)
    response_validated_count = sum(1 for r in test_results if r.get("response_match") is not None)
    response_match_count = sum(1 for r in test_results if r.get("response_match") is True)

    output = {
        "all_passed": all_passed,
        "passed_count": passed_count,
        "failed_count": failed_count,
        "total_count": total_count,
        "results": test_results,
        "failures": failures,
    }

    # Add contract validation summary if any tests had expected responses
    if response_validated_count > 0:
        output["contract_validation"] = {
            "tests_with_expected_response": response_validated_count,
            "response_matches": response_match_count,
            "response_mismatches": response_validated_count - response_match_count,
        }

    print("\n" + "=" * 60)
    print(f"Results: {passed_count}/{total_count} passed")
    if response_validated_count > 0:
        print(f"Contract validation: {response_match_count}/{response_validated_count} responses matched")
    print("=" * 60)
    print(json.dumps(output))
    sys.stdout.flush()  # Ensure output is flushed immediately
