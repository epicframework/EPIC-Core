import os
import zipfile
import yaml, json
import shutil
from flask import Flask, current_app, Blueprint, render_template, request, flash, redirect, url_for, Response
from werkzeug.utils import secure_filename

from bindings import new_ll, connect, new_interpreter

# Create local interpreter and remote interpreter
ll = new_ll("../connective/connective/sharedlib/elconn.so")
ll.elconn_init(1)
# TODO: get Connective URL from configuration
remote_connective = connect(ll, b"http://127.0.0.1:3111")
# remote_connective = connect(ll, b"http://127.0.0.1:3003")
connective = new_interpreter(ll)
# Allow messages to be send to remote interpreter by prefixing
# the command "hub"
ll.elconn_link(b"hub", connective.ii, remote_connective.ii)


control_panel = Blueprint('control_panel', __name__)

PACKAGE_DIRECTORY = './packages'
ALLOWED_EXTENSIONS = ['zip']

# Control Panel Main Page
@control_panel.route('/')
def index():
	# TODO: Check packages name matches spec name
	# TODO: Clean up invalid packages
	# Get Upload Directory
	UPLOAD_FOLDER = current_app.config['UPLOAD_FOLDER']
	# Get list of current packages
	packages_dir = os.listdir(UPLOAD_FOLDER)
	# Get all package specifications
	packages = []
	for dir_name in packages_dir:
		dir = os.path.join(UPLOAD_FOLDER, dir_name)
		if os.path.isdir(dir):
			# Check if valid directory is a package
			package_dir = os.listdir(dir)
			for file in package_dir:
				if file == "package.yml":
					package_config = os.path.join(dir, file)
					with open(package_config, "r") as stream:
						try:
							package = yaml.load(stream)
						except Exception as exc:
							package = exc
					packages.append(package)
	# Get Devices
	# Update local store
	# TODO: set a timer for this instead of doing it on every request
	connective.runs(': devices (store (hub devices list))')

	# Copy devices list from Connective to Python
	devices = connective.runs('devices', tolist=True)
	
	return render_template("index.html", packages = packages, devices = devices)
	


# View Packages
@control_panel.route('/packages')
def view_packages():
	# TODO: Check packages name matches spec name
	# TODO: Clean up invalid packages
	# Get Upload Directory
	UPLOAD_FOLDER = current_app.config['UPLOAD_FOLDER']
	# Get list of current packages
	packages_dir = os.listdir(UPLOAD_FOLDER)
	# Get all package specifications
	packages = []
	for dir_name in packages_dir:
		dir = os.path.join(UPLOAD_FOLDER, dir_name)
		if os.path.isdir(dir):
			# Check if valid directory is a package
			package_dir = os.listdir(dir)
			for file in package_dir:
				if file == "package.yml":
					package_config = os.path.join(dir, file)
					with open(package_config, "r") as stream:
						try:
							package = yaml.load(stream)
						except Exception as exc:
							package = exc
					packages.append(package)
	# Render View
	return render_template("view_packages.html", packages = packages)

# View Package
@control_panel.route('/packages/<package_name>')
def view_package(package_name):
	# Check for required paramter
	if package_name is not None:
		# Get Upload Directory
		UPLOAD_FOLDER = current_app.config['UPLOAD_FOLDER']
		# Get Package Directory
		package_dir = os.path.join(UPLOAD_FOLDER, package_name, "package.yml")
		# Get Package Config
		with open(package_dir, "r") as stream:
			try:
				package = yaml.load(stream)
				return render_template("view_package.html", package = package)
			except Exception as exc:
				exception = exc
				return render_template('error_template.html', exception = exception)
	# TODO: Throw actual exception
	else:
		return render_template("error_template.html", exception="Expected '/packages/<package_name>'. Got required parameter 'package_name' equal to None.")

# Helper method for add_package
def allowed_file(filename):
	return '.' in filename and \
		filename.rsplit('.', 1)[1].lower() in ALLOWED_EXTENSIONS

# Add Package
@control_panel.route('/packages/add', methods=['GET', 'POST'])
def add_package():
	if request.method == 'POST':
		if 'file' not in request.files:
			flash('Missing file')
			return render_template("add_package.html", error=request.files['file'])
		file = request.files['file']
		if file.filename == '':
			flash('No file selected')
			return render_template("add_package.html", error="No file selected")
		elif file and allowed_file(file.filename):
			filename = secure_filename(file.filename)
			file.save(os.path.join(current_app.config['UPLOAD_FOLDER'], filename))
			# TODO package set up
			return redirect('/packages/setup?package='+filename)

	return render_template("add_package.html")

# Setup Package
@control_panel.route('/packages/setup')
def setup_package():
	package = request.args.get('package')
	# Check for required parameter
	if package is not None:
		try:
			# Unzip package file
			# TODO: Prevent name collisions
			zip_ref = zipfile.ZipFile(os.path.join(current_app.config['UPLOAD_FOLDER'], package), mode='r')
			zip_ref = zip_ref.extractall(current_app.config['UPLOAD_FOLDER'])
			# Get Package instructions
			package_dir = package.rsplit('.')[0]
			config_dir = os.path.join(current_app.config['UPLOAD_FOLDER'], package_dir, 'package.yml')
			config = None
			with open(config_dir, "r") as stream:
				try:
					config = yaml.load(stream)
				except Exception as exc:
					return str(exc)

			# TODO: Add package instructions to jobs list

			# Inform subscriber about the new package
			connective.runl(["hub", "events", "new-package", "enque", config])

			return redirect('/packages')
		except Exception as exc:
			return render_template("error_template.html", exception=exc)
	# TODO: Throw real exception
	else:
		return render_template("error_template.html", exception="Expected query string 'package=<string>'. Got required query string 'package' equal to None.")

# Remove Package
@control_panel.route('/packages/remove/<package_name>')
def remove_package(package_name):
	# TODO Send kill command to Manager
	# Check for required paramter
	if package_name is not None:
		# Get Upload Directory
		UPLOAD_FOLDER = current_app.config['UPLOAD_FOLDER']
		package_dir = os.path.join(UPLOAD_FOLDER, package_name)
		# Remove Package
		try:
			shutil.rmtree(package_dir)
		except Exception as exc:
			return render_template("error_template.html", exception=exc)
		# Render success template
		return render_template("remove_package.html", removal=package_name)
	# TODO: Throw real exception
	else:
		return render_template("error_template.html", exception="Expected '/packages/remove/<package_name>'. Got required parameter 'package_name' equal to None.")

# View Devices
@control_panel.route('/devices')
def view_devices():
	# Update local store
	# TODO: set a timer for this instead of doing it on every request
	connective.runs(': devices (store (hub devices list))')

	# Copy devices list from Connective to Python
	devices = connective.runs('devices', tolist=True)

	return render_template("view_devices.html", devices = devices)

# View Device
@control_panel.route('/devices/<device_id>')
def view_device(device_id):
	# Update local store
	# TODO: set a timer for this instead of doing it on every request
	connective.runs(': devices (store (hub devices list))')

	# Copy devices list from Connective to Python
	devices = connective.runs('devices', tolist=True)

	device = None
	for dev in devices:
		if dev["internal_id"] == device_id:
			device = dev

	return render_template("view_device.html", device = device)

# Read Device Property
@control_panel.route('/devices/property/read', methods=['POST'])
def read_property():
	if request.method == 'POST':
		target = request.form.get('target')
		property_ = request.form.get('property')

		# Read property from device properties map (HA/Connective)
		result = connective.runl(
			'hub devices internal_registry'.split(' ') +
			[target, "properties", property_], tolist=True)
		return str(result[0])

# Write Device Property
@control_panel.route('/devices/property/write', methods=['POST'])
def write_property():
	if request.method == 'POST':
		target = request.form.get('target')
		property_ = request.form.get('property')
		value = request.form.get('value')

		propertyMap = {
			property_: value,
		}

		# Enqueue property update request (HA/Connective)
		hopefullyNotEqualToZero = connective.runl(
			'hub devices registry'.split(' ') +
			[target, "update-queue", "enque", propertyMap])
		return "placeholder"

# Trigger Device Action
@control_panel.route('/devices/action/trigger', methods=['POST'])
def trigger_action():
	if request.method == 'POST':
		target = request.form['target']
		action = request.form['action']

		# TODONE: Eric
		result = connective.runl(["hub", "devices", "internal_registry", target,
			"actions", action, "enque", True], tolist=True)
		return str(result)

# View Macros
#@control_panel.route('/macros')
#def view_macros():
#	macros = os.listdir('./macros')
#	macro_dict = {}
#	for macro in macros:
#		with open(macro, 'r')

# Add Macro
#@control_panel.route('/macros/add', methods=["POST"])
#def add_macro():
#	macro_name = request.form["macro_name"]
#	macro_content = request.form["macro_content"]
#	macro_path = os.path.join("./macros", macro_name)
#	with open(macro_path, "w") as stream:
#		try:
#			#TODO: write
#			pass
#		except Exception as exc:
#			return render_template("error_template.html", exception=exc)

# Remove Macro
#@control_panel.route('/macros/remove', methods=["POST"])
#def remove_macro():
#	macro_name = request.form["macro_name"]
#	macro_path = os.path.join("./macros", macro_name)
#	try:
#		#TODO: remove
#	except Exception as exc:
#		return render_template("error_template.html", exception=exc)

# Process Macro
#@control_panel.route('/macros/read', methdos=['POST'])
#def read_macro():
#	if request.method == 'POST':
#		macro = request.form['macro_name']
#		macro_path = os.path.join("./macros", macro)
#		with open(macro_path, 'r') as stream:
#			try:
#				macro_code = yaml.load(stream)
#			except Exception as exc:
#				return render_template("error_template.html", exception=exc)
#		vars = {}
#		response = []
#		for command in macro_code:
#			if command[0] == "read":
#				target = command[1]
#				property_ = command[2]
#				var_name = command[3]
#				result = connective.runl(
#					'hub devices internal_registry'.split(' ') +
#				[target, "properties", property_], tolist=True)
#				vals[var_name}
#			if command[0] == "print":
#				var_name = command[1]
#				response.append(vars[var_name])
#			else:
#				pass
