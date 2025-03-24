Rails.application.config.after_initialize do
  Rails.logger.info "hello from rails"
  puts "Hello from Rails" # This will show in the console/terminal
end
