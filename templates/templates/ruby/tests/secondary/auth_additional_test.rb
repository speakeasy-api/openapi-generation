# frozen_string_literal: true
# typed: true

require_relative "../lib/alphabetically_early"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "securerandom"

# Request log entry for recording HTTP requests made through Faraday
RequestLogEntry = Struct.new(:request_url, :request_body, :status_code, keyword_init: true)

# Faraday middleware that records requests while still making real HTTP calls
class RequestRecorderMiddleware < Faraday::Middleware
  attr_reader :log

  def initialize(app)
    super(app)
    @log = []
    @mutex = Mutex.new
  end

  def call(env)
    request_body = env.body.to_s
    response = @app.call(env)
    @mutex.synchronize do
      @log << RequestLogEntry.new(
        request_url: env.url.to_s,
        request_body: request_body,
        status_code: response.status
      )
    end
    response
  end
end

module AlphabeticallyEarly
  class TestAuthAdditional < Minitest::Test
    def test_authenticated_request_operation_level_oauth2
      record_test("auth-operation-level-oauth2")

      middleware = nil
      conn = Faraday.new do |f|
        f.use(RequestRecorderMiddleware)
        f.request(:multipart, { flat_encode: true })
        f.adapter Faraday.default_adapter
      end
      # Build the app to instantiate middleware instances, then walk the chain
      app = conn.builder.app
      while app
        if app.is_a?(RequestRecorderMiddleware)
          middleware = app
          break
        end
        app = app.respond_to?(:app) ? app.app : nil
      end
      recorder = middleware
      s = AlphabeticallyEarly::SDK.new(client: conn)
      refute_nil(s)

      client_secret = "supersecret-#{SecureRandom.hex(5)}"

      # A token should be requested with 'read', 'write' and 'erase' scopes
      s.hooks.authenticated_request(
        security: Models::Operations::AuthenticatedRequestSecurity.new(
          client_id: "speakeasy-sdks",
          client_secret: client_secret,
          audience: ""
        )
      )

      # This operation requires 'read' and 'write' scopes.
      # The same token should be reused since [read, write, erase] is a superset of [read, write].
      s.hooks.authenticated_request_unflattened(
        security: Models::Operations::AuthenticatedRequestUnflattenedSecurity.new(
          client_credentials: Models::Shared::SchemeClientCredentials.new(
            client_id: "speakeasy-sdks",
            client_secret: client_secret,
            audience: ""
          )
        )
      )

      # Verify that only a single token was requested
      token_requested = false
      recorder.log.each do |entry|
        next unless entry.request_url.include?("/clientcredentials/token")

        refute(token_requested, "Expected only a single token request")
        assert_includes(entry.request_body, "scope=read+write+erase")
        token_requested = true
      end
      assert(token_requested, "Expected at least one token request")
    end
  end
end
