# frozen_string_literal: true
# typed: true

require_relative "../lib/alphabetically_early"
require_relative "../lib/crystalline"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"
require "securerandom"

module AlphabeticallyEarly
  class TestRetriesAdditional < Minitest::Test
    def setup
      @sdk = AlphabeticallyEarly::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_retries_succeeds
      record_test("retries-succeeds")

      res = @sdk.retries.retries_get(request_id: SecureRandom.uuid)
      refute_nil(res)
      assert_equal(3, res.retries)
    end

    def test_retries_succeeds_with_body
      record_test("retries-succeeds-with-body")

      res = @sdk.retries.retries_post(
        request_id: SecureRandom.uuid,
        request_body: AlphabeticallyEarly::Models::Operations::RetriesPostRequestBody.new(field_one: "one")
      )
      refute_nil(res)
      assert_equal(3, res.retries)
    end

    def test_retries_request_timeout
      record_test("retries-request-timeout")
      error = nil
      begin
        @sdk.retries.retries_get_timeout(
          request_id: SecureRandom.uuid,
          fault_settings: "{\"delay_count\":1,\"delay_ms\":1500}",
          num_retries: 10,
          request_id_query_parameter: "retries-request-timeout",
          retries: nil,
          server_url: nil,
          timeout_ms: 1
        )
        # Faraday may raise either ConnectionFailed or TimeoutError for the
        # go-fault induced delay. Either which way though, this verifies that the
        # short timeout works.
      rescue Faraday::ConnectionFailed => e
        error = e
      rescue Faraday::TimeoutError => e
        error = e
      end

      refute_nil(error)
    end

    def test_retries_timeout
      record_test("retries-timeout")

      # NOTE: The raised error should change when proper response error handling added
      assert_raises(StandardError) do
        @sdk.retries.retries_get(
          request_id: SecureRandom.uuid,
          num_retries: 1_000_000_000,
          retries: AlphabeticallyEarly::Utils::RetryConfig.new(
            backoff: AlphabeticallyEarly::Utils::BackoffStrategy.new(
              exponent: 1.1,
              initial_interval: 1,
              max_elapsed_time: 100,
              max_interval: 50
            ),
            retry_connection_errors: false,
            strategy: "backoff"
          )
        )
      end
    end

    def test_retries_global_config_disable
      record_test("retries-global-config-disable")

      sdk = AlphabeticallyEarly::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        retry_config: AlphabeticallyEarly::Utils::RetryConfig.new(
          strategy: ""
        )
      )
      refute_nil(sdk)
      # NOTE: The raised error should change when proper response error handling added
      assert_raises(StandardError) do
        sdk.retries.retries_get(request_id: SecureRandom.uuid, num_retries: 2)
      end
    end

    def test_retries_global_config_success
      record_test("retries-global-config-success")

      sdk = AlphabeticallyEarly::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        retry_config: AlphabeticallyEarly::Utils::RetryConfig.new(
          backoff: AlphabeticallyEarly::Utils::BackoffStrategy.new(
            exponent: 1.1,
            initial_interval: 1,
            max_elapsed_time: 1000,
            max_interval: 50
          ),
          retry_connection_errors: false,
          strategy: "backoff"
        )
      )
      refute_nil(sdk)
      res = sdk.retries.retries_get(request_id: SecureRandom.uuid, num_retries: 20)
      refute_nil(res)
      assert_equal(20, res.retries)
    end

    def test_retries_global_config_timeout
      record_test("retries-global-config-timeout")

      sdk = AlphabeticallyEarly::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        retry_config: AlphabeticallyEarly::Utils::RetryConfig.new(
          backoff: AlphabeticallyEarly::Utils::BackoffStrategy.new(
            exponent: 1.1,
            initial_interval: 1,
            max_elapsed_time: 100,
            max_interval: 50
          ),
          retry_connection_errors: false,
          strategy: "backoff"
        )
      )
      refute_nil(sdk)
      # NOTE: The raised error should change when proper response error handling added
      assert_raises(StandardError) do
        sdk.retries.retries_get(request_id: SecureRandom.uuid, num_retries: 30)
      end
    end

    def test_retries_header
      record_test("retries-header")

      sdk = AlphabeticallyEarly::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        retry_config: AlphabeticallyEarly::Utils::RetryConfig.new(
          backoff: AlphabeticallyEarly::Utils::BackoffStrategy.new(
            exponent: 1.1,
            initial_interval: 5000,
            max_elapsed_time: 10_000,
            max_interval: 10_000
          ),
          retry_connection_errors: false,
          strategy: "backoff"
        )
      )
      refute_nil(sdk)
      res = sdk.retries.retries_after(request_id: SecureRandom.uuid, num_retries: 3, retry_after_val: 1)
      refute_nil(res)
      assert_equal(3, res.retries)
    end

    def test_retries_connect_error
      record_test("retries-connect-error")

      sdk = AlphabeticallyEarly::SDK.new(
        server_url: CommonHelpers.httpbin_url,
        retry_config: AlphabeticallyEarly::Utils::RetryConfig.new(
          backoff: AlphabeticallyEarly::Utils::BackoffStrategy.new(
            exponent: 1.1,
            initial_interval: 1,
            max_elapsed_time: 1000,
            max_interval: 50
          ),
          retry_connection_errors: true,
          strategy: "backoff"
        )
      )
      refute_nil(sdk)
      error = nil

      begin
        sdk.retries.retries_connect_error_get
      rescue Faraday::ConnectionFailed => e
        error = e
      end

      refute_nil(error)
    end
  end
end
