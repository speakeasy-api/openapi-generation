# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "../lib/crystalline"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"
require "json"
require "net/http"
require "time"
require "uri"

module OpenApiSDK
  # Faraday middleware that forces the first token request to return an expired token.
  # Used for testing token renewal behavior.
  class ExpireFirstTokenMiddleware < Faraday::Middleware
    def initialize(app, tracker)
      super(app)
      @tracker = tracker
    end

    def call(env)
      if env.method == :post && env.url.path == "/oauth2/token"
        @tracker[:count] += 1
        if @tracker[:count] <= 1
          env.request_headers["x-oauth2-expire-at"] = (Time.now - 86_400).iso8601
        end
      end
      @app.call(env)
    end
  end

  class TestOAuth2PasswordHook < Minitest::Test
    private

    def make_sdk_with_credentials(username: "testuser", password: "testpassword", client_id: "beezy", client_secret: "super-secret", **opts)
      OpenApiSDK::SDK.new(
        oauth2: Models::Shared::Oauth2Credentials.new(
          username: username,
          password: password,
          client_id: client_id,
          client_secret: client_secret
        ),
        **opts
      )
    end

    public

    def test_oauth2_password_with_credentials
      record_test("hooks-oauth2-password-with-credentials")

      s = make_sdk_with_credentials
      refute_nil(s)

      res = s.products.list
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal("pass", res.http_meta.response.headers["x-oauth2"])
    end

    def test_oauth2_password_with_token
      record_test("hooks-oauth2-password-with-token")

      # Manually obtain a token via direct HTTP request
      uri = URI("#{API_TEST_SERVICE_URL}/oauth2/token")
      http = Net::HTTP.new(uri.host, uri.port)
      request = Net::HTTP::Post.new(uri)
      request["Content-Type"] = "application/x-www-form-urlencoded"
      request.body = URI.encode_www_form(
        "grant_type" => "password",
        "username" => "testuser",
        "password" => "testpassword",
        "client_id" => "beezy",
        "client_secret" => "super-secret",
        "scope" => "products:read"
      )
      response = http.request(request)
      assert_equal("200", response.code)

      token_data = JSON.parse(response.body)
      token = token_data["access_token"]
      refute_nil(token)

      s = OpenApiSDK::SDK.new(
        oauth2: token
      )

      res = s.products.list
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal("pass", res.http_meta.response.headers["x-oauth2"])
    end

    def test_oauth2_password_token_renewal
      record_test("hooks-oauth2-password-token-renewal")

      tracker = { count: 0 }

      client = Faraday.new(url: API_TEST_SERVICE_URL) do |f|
        f.request :url_encoded
        f.use ExpireFirstTokenMiddleware, tracker
        f.adapter Faraday.default_adapter
      end

      s = make_sdk_with_credentials(client: client)

      # First request should fail with 401 because the token is expired
      error = assert_raises do
        s.products.list
      end
      refute_nil(error)

      # Second request should succeed because the session was cleared on 401
      res = s.products.list
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal("pass", res.http_meta.response.headers["x-oauth2"])
      assert_equal(2, tracker[:count])
    end

    def test_oauth2_password_operation_scope
      record_test("hooks-oauth2-password-operation-scope")

      s = make_sdk_with_credentials

      # listProducts requires scope "products:read"
      list_res = s.products.list
      refute_nil(list_res)
      assert_equal(Rack::Utils.status_code(:ok), list_res.http_meta.response.status)
      assert_equal("pass", list_res.http_meta.response.headers["x-oauth2"])

      # createProduct requires scope "products:create"
      create_res = s.products.create(
        request: Models::Shared::NewProductForm.new(
          description: "Games console",
          name: "Playstation 5",
          price: 499.99
        )
      )
      refute_nil(create_res)
      assert_equal(Rack::Utils.status_code(:ok), create_res.http_meta.response.status)
      assert_equal("pass", create_res.http_meta.response.headers["x-oauth2"])
    end

    def test_oauth2_password_bad_credentials
      record_test("hooks-oauth2-password-bad-credentials")

      s = make_sdk_with_credentials(password: "BAD_PASSWORD")

      error = assert_raises(RuntimeError) do
        s.products.list
      end
      assert_match(/invalid username or password/, error.message)
    end

    def test_oauth2_password_not_required
      record_test("hooks-oauth2-password-not-required")

      s = make_sdk_with_credentials

      res = s.health_check
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      # healthCheck has security: [] so no oauth2 header should be present
      x_oauth2 = res.http_meta.response.headers["x-oauth2"]
      assert(x_oauth2.nil? || x_oauth2.empty?, "Expected no x-oauth2 header for non-authenticated endpoint")
    end

    def test_oauth2_password_operation_security_option
      record_test("hooks-oauth2-password-operation-security-option")

      s = OpenApiSDK::SDK.new

      res = s.inventory.update_product_stock(
        security: Models::Operations::UpdateProductStockSecurity.new(
          oauth2: Models::Shared::Oauth2Credentials.new(
            username: "testuser",
            password: "testpassword",
            client_id: "beezy",
            client_secret: "super-secret"
          )
        ),
        product_inventory_update_form: Models::Shared::ProductInventoryUpdateForm.new(
          quantity_delta: 99
        ),
        id: "123"
      )
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      assert_equal("pass", res.http_meta.response.headers["x-oauth2"])
    end
  end
end
