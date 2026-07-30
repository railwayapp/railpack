set -eu

for path in \
	app/Justfile \
	app/README.md \
	app/TODO \
	negation_test/should_exist.txt \
	negation_test/existing_folder/file_in_folder.txt \
	start.sh \
	web/app.js
do
	if [ ! -e "$path" ]; then
		echo "expected included path to exist: $path" >&2
		exit 1
	fi
done

for path in \
	.env-specific \
	.envother \
	Justfile \
	README.md \
	__pycache__ \
	docker-compose-2.yml \
	docker-compose.yml \
	node_modules/root-package/index.js \
	the.log \
	tmp/.should-hide \
	tmp/should-hide \
	web/node_modules/some-package/index.js
do
	if [ -e "$path" ]; then
		echo "expected excluded path not to exist: $path" >&2
		exit 1
	fi
done

find . | sort
