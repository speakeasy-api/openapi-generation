# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "../lib/crystalline"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestErrorsAdditional < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_status_get_error_default_error_codes
      record_test("errors-status-get-error-default-error-codes")

      err = assert_raises(Models::Errors::APIError) do
        @sdk.errors.status_get_error(status_code: 400)
      end

      assert_equal(400, err.status_code)

      err = assert_raises(Models::Errors::APIError) do
        @sdk.errors.status_get_error(status_code: 500)
      end

      assert_equal(500, err.status_code)
    end

    def test_status_get_error_300_non_error
      record_test("errors-status-get-error300-non-error")

      res = @sdk.errors.status_get_error(status_code: 300)
      assert_equal(300, res.http_meta.response.status)
    end

    def test_status_get_error_x_speakeasy_errors
      record_test("errors-status-get-error-x-speakeasy-errors")

      err = assert_raises(Models::Errors::APIError) do
        @sdk.errors.status_get_x_speakeasy_errors(status_code: 400)
      end

      assert_equal(400, err.status_code)
      assert_equal("API error occurred", err.message)
    end

    def test_status_get_success_x_speakeasy_errors_empty_status_code_list
      record_test("errors-status-get-error-x-speakeasy-errors-none")

      res = @sdk.errors.status_get_non_error(status_code: 200)
      assert_equal(200, res.http_meta.response.status)

      res = @sdk.errors.status_get_non_error(status_code: 400)
      assert_equal(400, res.http_meta.response.status)

      res = @sdk.errors.status_get_non_error(status_code: 500)
      assert_equal(500, res.http_meta.response.status)
    end

    def test_status_get_success_x_speakeasy_errors_unspecified_responses
      record_test("errors-status-get-error-x-speakeasy-errors-default")

      res = @sdk.errors.status_get_default_error(status_code: 200)
      assert_equal(200, res.http_meta.response.status)

      res = @sdk.errors.status_get_default_error(status_code: 400)
      assert_equal(400, res.http_meta.response.status)

      err = assert_raises(Models::Errors::APIError) do
        @sdk.errors.status_get_default_error(status_code: 404)
      end
      assert_equal(404, err.status_code)

      err = assert_raises(Models::Errors::APIError) do
        @sdk.errors.status_get_default_error(status_code: 500)
      end
      assert_equal(500, err.status_code)

      # To make sure the catch-all "default" code gets properly applied in `templateErrorStatusCodesCheck`,
      # an AfterError hook was added to test_hook.rb to recover from 418 error.
      res = @sdk.errors.status_get_default_error(status_code: 418)
      assert_equal(200, res.http_meta.response.status)
    end

    def test_connection_error_get
      record_test("errors-connection-error")
      err = assert_raises(StandardError) do
        @sdk.errors.connection_error_get
      end

      assert(err.message.start_with?("Failed to open TCP connection"))
    end

  end
end
