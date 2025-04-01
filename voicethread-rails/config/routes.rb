Rails.application.routes.draw do
  get 'webhooks/stripe'
  get 'subscriptions/new'
  get 'subscriptions/create'
  get 'subscriptions/success'
  get 'subscriptions/cancel'
  # Define your application routes per the DSL in https://guides.rubyonrails.org/routing.html

  # Defines the root path route ("/")
  root "pages#home"

  # Devise routes
  devise_for :users

  # Pages routes
  get 'about', to: 'pages#about'
  get 'features', to: 'pages#features'
  get 'pricing', to: 'pages#pricing'

  # Topics routes
  resources :topics

  # Stripe routes
  resources :subscriptions, only: [:new, :create] do
    collection do
      get :success
      get :cancel
    end
  end

  # Webhook route for Stripe
  post 'webhooks/stripe', to: 'webhooks#stripe'

  # Action Cable WebSocket routes
  mount ActionCable.server => '/cable'

  # Reveal health status on /up that returns 200 if the app boots with no exceptions, otherwise 500.
  # Can be used by load balancers and uptime monitors to verify that the app is live.
  get "up" => "rails/health#show", as: :rails_health_check
end
