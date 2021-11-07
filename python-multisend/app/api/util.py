import binascii
import os
import time
import uuid
import json

import app.node
from app.api.models import PrivateAddress
import app.api.models as models
from web3 import Web3

chainId = int(os.getenv("ETH_CHAIN_ID", 3))


class Multisend(object):
    def __init__(self, session: app.node.Node):
        # self.storage = storage
        self.session = session
        self.w3 = session.w3

        self.bank = ""

    def getBankAddress(self):
        return self.bank

    def setBankAddress(self, key):
        self.bank = PrivateAddress.addPrivateKey(private=key)

    def validateTransaction(self, tx):
        valid = {
            "from": tx['from'],
            "nonce": tx['nonce'],
            "gasPrice": tx['gasPrice'],
            "gas": self.w3.toHex(tx['gas']),
            "to": tx['to'],
            "value": '0x0',
            "chainId": chainId
        }
        for i in ['data', 'value']:
            if tx.get(i):
                valid[i] = tx.get(i)

        return valid

    def PreparePayments(self, data, token_address):
        # collect tx to spend group
        walletSpendGroup = {}
        for tx in data:
            if walletSpendGroup.get(tx['from']):
                walletSpendGroup[tx['from']]['value'] += tx['amount']
                walletSpendGroup[tx['from']]['list'].append(tx)
            else:
                walletSpendGroup[tx['from']] = {'value': tx['amount'], 'list': [tx]}

        # check allowance
        allowCheck = [[wallet, walletSpendGroup[wallet]['value']] for wallet in walletSpendGroup]
        txRequire = []
        setMaxApprovePolice = True
        for i in allowCheck:
            allowed = self.session.getAllowance(token_address, i[0])
            if allowed < i[1]:
                # we cant rewrite approve -> we must set it to 0 then write new value
                nonceINC = 0
                if allowed > 0:
                    approve = self.session.getApprove(token_address, i[0], 0)
                    txRequire.append(approve)
                    nonceINC = 1

                approve = self.session.getApprove(token_address, i[0], -1 if setMaxApprovePolice else i[1], nonceINC=nonceINC)
                txRequire.append(approve)

        # check is address balances enough to confirm
        # calculate require gas to finish
        walletSpendGroup = {}
        for tx in txRequire:
            print (int(tx['gasPrice'], base=16))
            if walletSpendGroup.get(tx['from']):
                walletSpendGroup[tx['from']]['gas'] += tx['gas'] * int(tx['gasPrice'], base=16)
            else:
                walletSpendGroup[tx['from']] = {'gas': tx['gas'] * int(tx['gasPrice'], base=16)}

        requireBalance = [[wallet, walletSpendGroup[wallet]['gas']] for wallet in walletSpendGroup]
        addBalance = []
        for wallet in requireBalance:
            balance = self.session.getETHBalance(wallet[0])
            diff = wallet[1] - balance
            if diff > 0:
                addBalance.append([wallet[0], diff])
                # addBalance.append(self.session.getSendEth(self.getBankAddress(), wallet[0], diff))

        return {"tx_require": txRequire, "add_balance": addBalance}

    def GetEstimated(self, paymentData, inputList):
        gas = 0
        current_gas_price = self.w3.eth.gasPrice

        if len(paymentData['add_balance']) > 0:
            for i in paymentData['add_balance']:
                tx = self.session.getSendEth(self.getBankAddress(), i[0], i[1], current_gas_price)
                gas += tx['gas'] * int(tx['gasPrice'], base=16)
                gas += i[1]

        if len(inputList) > 0:
            gas += (40000 + (15000 * len(inputList))) * current_gas_price

        return {"gas": int(gas / current_gas_price), "current_gas_price": current_gas_price, "total_eth": "{:.8f}".format(self.session.WAI2ETH(gas))}
        # return {"gas": int(gas / current_gas_price), "current_gas_price": current_gas_price, "total_eth": self.session.WAI2ETH(gas)}


    def GetMultiSend(self, data):
        return

    def ConfirmPayments(self, paymentData, inputData, token_address, decimals):
        chain = []
        if len(paymentData['add_balance']) > 0:
            addBalance = []
            for i in paymentData['add_balance']:
                addBalance.append(self.session.getSendEth(self.getBankAddress(), i[0], i[1]))
            chain.append(addBalance)

        while len(paymentData['tx_require']) > 0:
            payments = []
            next = []
            lastFrom = ""
            while len(paymentData['tx_require']) > 0:
                value = paymentData['tx_require'].pop(0)

                if value['from'] == lastFrom:
                    next.append(value)
                    continue
                lastFrom = value['from']

                payments.append(value)
            paymentData['tx_require'] = next

            chain.append(payments)

        chain = [{"status": "wait", "data": event} for event in chain]

        chain.append({
            "status": "wait",
            "data": [self.session.getMultiSend(inputData, token_address, decimals)],
            "options": {"refresh_gas": True, "refresh_nonce": True, "refresh_limit": True}
        })

        if len(chain) == 0:
            return None

        record = models.MultiSendChain()
        record.input = inputData
        record.log = []
        record.data = chain
        record.insert()
        return record

    def SendTransaction(self, tx, options=None):
        if options is None:
            options = {}

        defaultOptions = {
            'refresh_gas': False,
            'refresh_limit': False,
            'refresh_nonce': False,
        }
        for i in options:
            defaultOptions[i] = options[i]

        count = self.w3.eth.getTransactionCount(Web3.toChecksumAddress(tx['from']))
        tx['nonce'] = self.w3.toHex(count)
        private = PrivateAddress.getPrivateKey(tx['from'])

        if defaultOptions['refresh_gas']:
            tx['gasPrice'] = self.w3.toHex(self.w3.eth.gasPrice)

        gasConfirm = True
        if defaultOptions['refresh_gas']:
            gasConfirm = False
            prev_gas = tx['gas']
            try:
                tx['gasLimit'] = prev_gas
                del tx['gas']
                tx['gas'] = self.w3.eth.estimateGas(tx)
                del tx['gasLimit']
                gasConfirm = True
            except Exception as e:
                tx['gas'] = prev_gas
                if tx.get('gasLimit'):
                    del tx['gasLimit']
                print('cant send transaction', e)

            # tx['gas'] = self.w3.eth.estimateGas(tx)

        if defaultOptions['refresh_nonce']:
            count = self.w3.eth.getTransactionCount(Web3.toChecksumAddress(tx['from']))
            tx['nonce'] = self.w3.toHex(count)

        if not gasConfirm:
            return

        validTransaction = self.validateTransaction(tx)
        signed_message = self.w3.eth.account.sign_transaction(validTransaction, private_key=private)
        result = self.w3.eth.send_raw_transaction(signed_message.rawTransaction)
        watch_hash = '0x' + binascii.hexlify(result).decode("utf-8")

        print(watch_hash)

        record = models.MultiSendWatchTransaction(watch_hash, "ETH")
        record.id = tx['id']
        record.insert()

    def CheckChain(self, chain, log, transactions, lastBlock):
        Save = False
        for event in chain:
            if event.get('status') == "done":
                continue

            pendingCount = len(event.get('data'))
            for tx in event.get('data'):
                if tx.get('id') is None:
                    tx['id'] = uuid.uuid4().hex
                    # Send transaction
                    try:
                        self.SendTransaction(tx, options=event.get('options'))
                        log.append({"tx": tx, "time": int(time.time())})
                    except Exception as e:
                        log.append({"error": str(e), "tx": tx, "time": int(time.time())})
                        return [True, chain, log]
                    Save = True
                else:
                    pending = transactions.get(uuid.UUID(tx['id']))
                    if pending is None:
                        continue

                    if pending.status == 'valid':
                        if pending.info.get('status') == 0:
                            event['status'] = "invalid"
                            log.append({"status": "invalid", "time": int(time.time())})
                            Save = True
                            break
                        else:
                            pendingCount -= 1
                    elif pending.status == 'invalid':
                        event['status'] = "invalid"
                        log.append({"status": "invalid", "time": int(time.time())})
                        Save = True
                        break

            if pendingCount <= 0:
                event['status'] = "done"
                log.append({"status": "invalid", "time": int(time.time())})
                Save = True
                continue

            break

        print(Save, json.dumps(chain, indent=4, sort_keys=True))
        return [Save, chain, log]
