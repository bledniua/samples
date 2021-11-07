# Import flask dependencies
import os
import __main__

from flask import Blueprint, request, jsonify

# Import the database object from the main app module
from app import db

# Import module models
from app.api.models import PrivateAddress

from app.api.util import Multisend
from app.node import session

mod_api = Blueprint('api', __name__, url_prefix='/api')

unit = Multisend(session=session)


if __main__.__file__ is "run.py":
    unit.setBankAddress(os.getenv("PRIVATE_ROOT"))
    print("getBankAddress", unit.getBankAddress())


# Set the route and accepted methods
@mod_api.route('/prepare', methods=["POST"])
def prepare():
    body = request.get_json()
    paymentInfo = unit.PreparePayments(body['list'], body['token'])
    response = {}
    response['list'] = body['list']
    response['token'] = body['token']
    response['payment'] = paymentInfo
    response['estimated'] = unit.GetEstimated(paymentInfo, body['list'])

    return jsonify(response)


@mod_api.route('/confirm', methods=["POST"])
def confirm():
    body = request.get_json()
    # paymentData, inputData, token_address, decimals
    response = unit.ConfirmPayments(body['payment'], body['list'], body['token'], 1000000)
    # print(response)
    return jsonify({"input": response.input, "id": response.id})


@mod_api.route('/add_address', methods=["POST"])
def add_address():
    body = request.get_json()
    return jsonify(PrivateAddress.addPrivateKey(body['key']))


# @mod_api.route('/get_address', methods=["POST"])
# def get_address():
#     body = request.get_json()
#     return jsonify(PrivateAddress.getPrivateKey(body['key']))

@mod_api.route('/address_list', methods=["GET"])
def get_address():
    return jsonify(PrivateAddress.getAllPublicKey())
