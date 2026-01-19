# We check the memory map of the current process to see if the
# shared library is actually loaded.
if File.read("/proc/self/maps").include?("libjemalloc")
  puts "jemalloc available"
else
  puts "jemalloc not available"
end
