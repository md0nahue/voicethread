class DeviseMailer < Devise::Mailer
  helper :application
  include Devise::Controllers::UrlHelpers

  default template_path = 'devise/mailers'
  default from = ENV['AWS_SES_FROM_EMAIL']

  def confirmation_instructions(record, token, opts={})
    @token = token
    super
  end

  def reset_password_instructions(record, token, opts={})
    @token = token
    super
  end

  def email_changed(record, opts={})
    super
  end

  def password_change(record, opts={})
    super
  end

  def unlock_instructions(record, token, opts={})
    @token = token
    super
  end

  private

  def headers_for(action, opts)
    super.merge(
      'X-SES-CONFIGURATION-SET' => ENV['AWS_SES_CONFIGURATION_SET']
    )
  end
end 