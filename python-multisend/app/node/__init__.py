import os
import os.path
from decimal import Decimal

from app.node import method

from web3 import Web3

chainId = os.getenv("CHAIN_ID", "0x03")

# decimals, contractAddress, contractABI, spenderAddress
class Node(object):
    def __init__(self, nodeAddress, spenderAddress):
        self.w3 = Web3(Web3.HTTPProvider(nodeAddress))
        self.abiCache = {}
        self.spenderAddress = spenderAddress



    def WAI2ETH(self, wai):
        return wai / self.EthInWai()

    def EthInWai(self):
        return 1000000000000000000

    def getETHBalance(self, address):
        return self.w3.eth.getBalance(Web3.toChecksumAddress(address))

    def _getAbi(self, contractAddress):
        if self.abiCache.get(contractAddress) is None:
            path = "abi/{}.json".format(contractAddress)
            if not os.path.isfile(path):
                raise Exception(path + " not found")

            with open(path, "r") as f: abi = f.read()
            self.abiCache[contractAddress] = abi

        return self.abiCache.get(contractAddress)

    def _getContract(self, contractAddress):
        return self.w3.eth.contract(Web3.toChecksumAddress(contractAddress), abi=self._getAbi(contractAddress))

    def getAllowance(self, contractAddress, address):
        contract = self._getContract(contractAddress)
        decimals = contract.functions.decimals().call()

        allowed = contract.functions.allowance(Web3.toChecksumAddress(address),
                                               Web3.toChecksumAddress(self.spenderAddress)).call()

        return allowed / Decimal(10 ** decimals)

    def getApprove(self, contractAddress, address, value, nonceINC=0):
        contract = self._getContract(contractAddress)
        decimals = contract.functions.decimals().call()

        count = self.w3.eth.getTransactionCount(Web3.toChecksumAddress(address)) + nonceINC

        gasPrice = self.w3.eth.gasPrice
        gasLimit = 50000
        if value == 0:
            gasLimit = 32000

        raw = {
            "from": Web3.toChecksumAddress(address),
            "nonce": self.w3.toHex(count),
            "gasPrice": self.w3.toHex(gasPrice),
            "gas": gasLimit,
            "to": Web3.toChecksumAddress(contractAddress),
            "value": "0x0",
            "data": method.approve(Web3.toChecksumAddress(self.spenderAddress), value, decimals),
        }
        try:
            raw['gasLimit'] = gasLimit
            del raw['gas']
            raw['gas'] = self.w3.eth.estimateGas(raw)
            del raw['gasLimit']
        except Exception as e:
            raw['gas'] = gasLimit
            print(e)

        return raw

    def getSendEth(self, bankAccount, address, value, gasPrice=0):
        count = self.w3.eth.getTransactionCount(Web3.toChecksumAddress(bankAccount))

        if gasPrice == 0:
            gasPrice = self.w3.eth.gasPrice
        gasLimit = 29000

        raw = {
            "from": Web3.toChecksumAddress(bankAccount),
            "nonce": self.w3.toHex(count),
            "gasPrice": self.w3.toHex(gasPrice),
            "gas": self.w3.toHex(gasLimit),
            "to": Web3.toChecksumAddress(address),
            "value": self.w3.toHex(value)
        }
        estimated = self.w3.eth.estimateGas(transaction=raw)
        raw['gas'] = estimated

        return raw

    def getContractOwner(self, address):
        contract = self._getContract(address)
        return contract.functions.owner().call()

    def getMultiSend(self, tx_list, token_address, decimals):
        ownerAddress = self.getContractOwner(self.spenderAddress)

        count = self.w3.eth.getTransactionCount(Web3.toChecksumAddress(ownerAddress))

        gasPrice = self.w3.eth.gasPrice
        gasLimit = 300000

        spender = []
        receiver = []
        value = []
        for i in tx_list:
            spender.append(i['from'])
            receiver.append(i['to'])
            value.append(i['amount'])

        raw = {
            "from": Web3.toChecksumAddress(ownerAddress),
            "nonce": self.w3.toHex(count),
            "gasPrice": self.w3.toHex(gasPrice),
            "gas": self.w3.toHex(gasLimit),
            "to": Web3.toChecksumAddress(self.spenderAddress),
            "value": '0x0',
            "data": method.multisend(token_address, spender, receiver, value, decimals),
        }

        return raw


MULTI_SEND_CONTRACT = os.getenv('MULTI_CONTRACT')
ETH_NODE_URL = os.getenv('ETH_NODE_URL', '')

session = Node(spenderAddress=MULTI_SEND_CONTRACT, nodeAddress=ETH_NODE_URL)
