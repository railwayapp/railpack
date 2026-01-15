---
title: Python
description: Building Python applications with Railpack
---

Railpack builds and deploys Python applications with support for various package managers and dependency management tools.

## Detection

Your project will be detected as a Python application if any of these conditions are met:

- One of `main.py`, `app.py`, `start.py`, `bot.py`, `hello.py`, or `server.py` exists in the root directory
- A `requirements.txt` file exists
- A `pyproject.toml` file exists
- A `Pipfile` exists

## Versions

Railpack supports Python 3.10 and later. We only officially support Python
versions that are actively maintained by the Python Software Foundation (not
EOL). See [Python release status](https://endoflife.date/python) for current
support status.

The Python version is determined in the following order:

- Set via the `RAILPACK_PYTHON_VERSION` environment variable
- Read from mise-compatible version files (`.python-version`,
  `.tool-versions`, `mise.toml`)
- Read from the `runtime.txt` file
- Read from the `Pipfile` if present
- Defaults to `3.13.2`

## Runtime Variables

These variables are available at runtime:

```sh
PYTHONFAULTHANDLER=1
PYTHONUNBUFFERED=1
PYTHONHASHSEED=random
PYTHONDONTWRITEBYTECODE=1
PIP_DISABLE_PIP_VERSION_CHECK=1
PIP_DEFAULT_TIMEOUT=100
```

## Configuration

Railpack builds your Python application based on your project structure. The build process:

- Installs Python and required system dependencies
- Installs project dependencies using your preferred package manager
- Configures the Python environment for production

The start command is determined by:

1. Framework specific start command (see below)
2. Main Python file in the root directory (checked in order: `main.py`,
   `app.py`, `start.py`, `bot.py`, `hello.py`, `server.py`)

### Package Managers

Railpack supports multiple Python package managers:

- **pip** - Uses `requirements.txt` for dependencies
- **poetry** - Uses `pyproject.toml` and `poetry.lock`
- **pdm** - Uses `pyproject.toml` and `pdm.lock`
- **uv** - Uses `pyproject.toml` and `uv.lock`
- **pipenv** - Uses `Pipfile`

### Config Variables

| Variable                   | Description                 | Example      |
| -------------------------- | --------------------------- | ------------ |
| `RAILPACK_PYTHON_VERSION`  | Override the Python version | `3.11`       |
| `RAILPACK_DJANGO_APP_NAME` | Django app name             | `myapp.wsgi` |

### System Dependencies

Railpack installs system dependencies for common Python packages:

- **pycairo**: Installs `libcairo2-dev` (build time) and `libcairo2` (runtime)
- **pdf2image**: Installs `poppler-utils`
- **pydub**: Installs `ffmpeg`
- **pymovie**: Installs `ffmpeg`, `qt5-qmake`, and related Qt packages

## Framework Support

Railpack detects and configures start commands for popular frameworks:

### FastHTML

Railpack detects FastHTML projects when `python-fasthtml` is listed as a
dependency. When detected:

- Starts with `uvicorn main:app --host 0.0.0.0 --port ${PORT:-8000}` if
  `uvicorn` is a dependency

### Flask

Railpack detects Flask projects when `flask` is listed as a dependency. When
detected:

- Starts with `gunicorn --bind 0.0.0.0:${PORT:-8000} main:app` if `gunicorn`
  is a dependency

### FastAPI

Railpack detects FastAPI projects when `fastapi` is listed as a dependency.
When detected and `uvicorn` is available as a dependency:

- Starts with `uvicorn main:app --host 0.0.0.0 --port ${PORT:-8000}`

### Django

Railpack detects Django projects by:

- Presence of `manage.py`
- Django being listed as a dependency

The start command is determined by:

1. `RAILPACK_DJANGO_APP_NAME` environment variable
2. Scanning Python files for `WSGI_APPLICATION` setting
3. Runs `python manage.py migrate && gunicorn {appName}:application`

### Databases

Railpack automatically installs system dependencies for common databases:

- **PostgreSQL**: Installs `libpq-dev` at build time and `libpq5` at runtime
- **MySQL**: Installs `default-libmysqlclient-dev` at build time and `default-mysql-client` at runtime

## BuildKit Caching

The Python provider will cache `/opt/pip-cache` under the cache key `pip`, and, for `uv`-based apps, `/opt/uv-cache` under the key `uv`.