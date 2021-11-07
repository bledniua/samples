# Import flask and template operators
from flask import Flask, jsonify

# Import SQLAlchemy
from flask_sqlalchemy import SQLAlchemy
import redis
import os
from dotenv import load_dotenv
load_dotenv()

REDIS_HOST = os.getenv('REDIS_HOST', 'localhost')
REDIS_PORT = os.getenv('REDIS_PORT', 6379)
REDIS_DB = os.getenv('REDIS_DB', 0)

# Define the WSGI application object
app = Flask(__name__)


# Sample HTTP error handling
@app.errorhandler(404)
def not_found(error):
    return jsonify({"error": "method not found"})


# Configurations
app.config.from_object('config')

# Define the database object which is imported
# by modules and controllers
db = SQLAlchemy(app)
r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, db=REDIS_DB)


# Register blueprint(s)
from app.api.controllers import mod_api as api

app.register_blueprint(api)
# app.register_blueprint(xyz)
# ..

# Build the database:
# This will create the database file using SQLAlchemy
db.create_all()