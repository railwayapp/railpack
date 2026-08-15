# The Ruby and Node deploy steps overlap, which currently duplicates inherited PATH entries.
# Assert the runtime value so this example reproduces the image configuration bug.
paths = ENV.fetch("PATH").split(":")
duplicates = paths.tally.filter_map { |path, count| path if count > 1 }

abort "Duplicate PATH entries: #{duplicates.join(", ")}" unless duplicates.empty?

puts "PATH contains no duplicate entries"
