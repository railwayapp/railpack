Rails.application.routes.draw do
  get "/" => lambda { |_env| [200, {}, ["Hello from Ruby: #{RUBY_VERSION}"]] }
end
