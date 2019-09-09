import datetime

import flask

import common

CACHE_TIMEOUT = datetime.timedelta(days=365).total_seconds()

app = flask.Flask(__name__, static_folder=None)


@app.route('/', defaults={'path': ''})
@app.route('/<path:path>')
def serve_file(path):
  query = flask.request.query_string.decode()
  mimetype = common.content_type(path, query)
  path = common.make_path(path, query)
  return flask.send_from_directory('files', path, mimetype=mimetype,
                                   cache_timeout=CACHE_TIMEOUT)


if __name__ == '__main__':
  app.run('0.0.0.0', 8000)
