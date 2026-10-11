set -eu

if [ ! -f kept.txt ]; then
	echo "kept.txt should survive the railpack.json negation" >&2
	exit 1
fi

for path in secret.txt also-secret.txt; do
	if [ -e "$path" ]; then
		echo "expected $path to be excluded" >&2
		exit 1
	fi
done

echo "kept.txt"
echo "secret.txt excluded"
echo "also-secret.txt excluded"
