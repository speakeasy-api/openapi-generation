# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module OpenApiSDK
  class TestHeadersAdditional < Minitest::Test
    def test_headers_override_request_headers
      record_test("headers-override-request-headers")

      sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(sdk)

      res = sdk.methods.method_get(
        http_headers: {
          "x-inject-header-1" => "foo",
          "x-inject-header-2" => "bar"
        }
      )

      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.request)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
      refute_nil(res.object)
      assert_equal("OK", res.object.status)
      assert_equal("foo", res.http_meta.request.headers["x-inject-header-1"])
      assert_equal("bar", res.http_meta.request.headers["x-inject-header-2"])
    end

    def test_pagination_keeps_http_headers_on_next_page
      sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
      refute_nil(sdk)

      responses = sdk.pagination.pagination_limit_offset_optional_page_params(
        page: 0,
        http_headers: {
          "x-pagination-header" => "keep-me"
        }
      )

      count = 0
      responses.each do |response|
        break if count == 2
        refute_nil(response)
        refute_nil(response.http_meta)
        refute_nil(response.http_meta.request)
        refute_nil(response.http_meta.response)
        assert_equal(Rack::Utils.status_code(:ok), response.http_meta.response.status)
        assert_equal("keep-me", response.http_meta.request.headers["x-pagination-header"])
        assert_equal(count.to_s, response.http_meta.request.params["page"])
        count += 1
      end

      assert_equal(2, count)
    end
  end
end
