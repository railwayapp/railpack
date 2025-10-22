Rails.application.routes.draw do
  get "/" => lambda { |_env| [200, {}, ["OK"]] }
end
