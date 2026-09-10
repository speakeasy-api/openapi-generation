# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"

class TestBasicAuthOptional < Minitest::Test
  def test_basic_auth_operation_optional
    record_test("auth-basic-auth-operation-optional")

    sdk = OpenApiSDK::SDK.new(
      security: OpenApiSDK::Models::Components::Security.new(
        username: "wrongUser",
        password: "wrongPass"
      )
    )

    res = sdk.auth.basic_auth_optional(
      security: OpenApiSDK::Models::Operations::BasicAuthOptionalSecurity.new(
        username: "testUser",
        password: "testPass"
      )
    )

    refute_nil(res)
    assert_equal(200, res.status_code)
    refute_nil(res.basic_auth_response)
    assert(res.basic_auth_response.authenticated)
    assert_equal("testUser", res.basic_auth_response.user)


    sdk2 = OpenApiSDK::SDK.new(
      security: OpenApiSDK::Models::Components::Security.new(
        username: "testUser",
        password: "testPass"
      )
    )

    assert_raises do
      sdk2.auth.basic_auth_optional(
        security: OpenApiSDK::Models::Operations::BasicAuthOptionalSecurity.new
      )
    end
  end
end
