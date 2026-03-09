if RubyVM::YJIT.enabled?
  yjit = "enabled"
else
  yjit = "not enabled"
end
puts "Hello from Ruby #{RUBY_VERSION}! YJIT is #{yjit}."
puts "Ruby version: #{RUBY_VERSION}"
