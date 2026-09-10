# frozen_string_literal: true
# typed: true

require_relative "../lib/alphabetically_early"
require_relative "../lib/crystalline"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module AlphabeticallyEarly
  class TestErrorsAdditional < Minitest::Test
    def setup
      @sdk = AlphabeticallyEarly::SDK.new(server_url: CommonHelpers.httpbin_url)
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

      @sdk.errors.status_get_error(status_code: 300)
    end

    def test_status_get_error_x_speakeasy_errors
      record_test("errors-status-get-error-x-speakeasy-errors")

      err = assert_raises(Models::Errors::APIError) do
        @sdk.errors.status_get_x_speakeasy_errors(status_code: 400)
      end

      assert_equal(400, err.status_code)
      assert_equal("API error occurred", err.message)
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
