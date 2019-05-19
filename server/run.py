from flask import Flask, send_from_directory
from control_panel.views import control_panel

# TODO: Configure upload folder
UPLOAD_FOLDER = './packages'

def create_app(config_file):
	app = Flask(__name__) # Create application object

	@app.route('/js/<path:path>')
	def send_js(path):
		return send_from_directory('js', path)

	@app.route('/css/<path:path>')
	def send_css(path):
		return send_from_directory('css', path)

	app.config.from_pyfile(config_file) # Configure application
	app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER
	app.register_blueprint(control_panel) # Register control panel
	app.config['SESSION_TYPE'] = 'memcached'
	app.config['SECRET_KEY'] = 'super secret key'
	return app


if __name__ == '__main__':
	app = create_app('config.py') # Create application with config.py
	app.run(host='0.0.0.0', debug=True) # Run Flask application
