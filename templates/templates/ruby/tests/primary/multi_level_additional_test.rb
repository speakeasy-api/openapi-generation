# frozen_string_literal: true
# typed: true

require_relative "../lib/openapi"
require_relative "common_helper_test"
require_relative "helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"
require "active_support"

module OpenApiSDK
  class TestMultiLevel < Minitest::Test
    def setup
      @sdk = OpenApiSDK::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_multi_level_grouping
      record_test("multi-level-grouping")

      assert_instance_of(OpenApiSDK::SDK, @sdk)
      res = @sdk.nested.first.get
      refute_nil(res)
      refute_nil(res.http_meta)
      refute_nil(res.http_meta.response)
      assert_equal(Rack::Utils.status_code(:ok), res.http_meta.response.status)
    end
  end
end
