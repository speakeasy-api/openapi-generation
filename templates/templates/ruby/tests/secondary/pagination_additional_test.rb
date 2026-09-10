# frozen_string_literal: true
# typed: true

require_relative "../lib/alphabetically_early"
require_relative "common_helper_test"

require "minitest/autorun"
require "minitest/focus"
require "rack"

module AlphabeticallyEarly
  class TestPagination < Minitest::Test
    def setup
      @sdk = AlphabeticallyEarly::SDK.new(server_url: CommonHelpers.httpbin_url)
    end

    def test_pagination_limit_offset_page_params_flat
      record_test("pagination-limit-offset-page-params-flat")

      assert_instance_of(AlphabeticallyEarly::SDK, @sdk)

      responses = @sdk.pagination.pagination_limit_offset_page_params(page: 1)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.result)
          assert_equal(20, response.result.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.result)
          assert_equal(0, response.result.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end

    def test_pagination_limit_offset_union_output_page_params_flat
      record_test("pagination-limit-offset-union-output-page-params-flat")

      assert_instance_of(AlphabeticallyEarly::SDK, @sdk)

      server_limit = 20
      responses = @sdk.pagination.pagination_limit_offset_union_output_page_params(page: 1)

      count = 0
      responses.each do |response|
        if count.zero?
          assert(response)
          assert(response.result)
          assert_equal(server_limit, response.result.result_array.size)
        elsif count == 1
          assert(response)
          assert(response.result)
          assert_equal(0, response.result.result_array.size)
        elsif count == 2
          assert_nil(response)
        else
          flunk("Unexpected response")
        end

        count += 1
      end
    end
  end
end
