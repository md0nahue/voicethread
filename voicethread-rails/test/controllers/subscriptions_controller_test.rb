require "test_helper"

class SubscriptionsControllerTest < ActionDispatch::IntegrationTest
  test "should get new" do
    get subscriptions_new_url
    assert_response :success
  end

  test "should get create" do
    get subscriptions_create_url
    assert_response :success
  end

  test "should get success" do
    get subscriptions_success_url
    assert_response :success
  end

  test "should get cancel" do
    get subscriptions_cancel_url
    assert_response :success
  end
end
