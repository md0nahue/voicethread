require "test_helper"

class PagesControllerTest < ActionDispatch::IntegrationTest
  test "should get home" do
    get pages_home_url
    assert_response :success
  end

  test "should get about" do
    get pages_about_url
    assert_response :success
  end

  test "should get features" do
    get pages_features_url
    assert_response :success
  end

  test "should get pricing" do
    get pages_pricing_url
    assert_response :success
  end
end
