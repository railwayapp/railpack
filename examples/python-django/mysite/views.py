from django.http import HttpResponse
from django.core.cache import cache
import redis
import os


def home(request):
    cache.set('test_key', 'Hello from Redis!', 30)
    cached_value = cache.get('test_key')

    redis_url = os.environ.get('REDIS_URL', 'redis://localhost:6379/0')
    r = redis.from_url(redis_url)
    redis_info = r.info()

    try:
        import hiredis
        hiredis_available = True
        hiredis_version = hiredis.__version__
    except ImportError:
        hiredis_available = False
        hiredis_version = 'not installed'

    response = f"""
    <html>
    <body>
        <h1>Django + PostgreSQL + Redis</h1>
        <p>Cached value from Redis: {cached_value}</p>
        <p>Redis version: {redis_info.get('redis_version', 'unknown')}</p>
        <p>Redis connected clients: {redis_info.get('connected_clients', 'unknown')}</p>
        <p>Hiredis available: {hiredis_available}</p>
        <p>Hiredis version: {hiredis_version}</p>
    </body>
    </html>
    """
    return HttpResponse(response)
