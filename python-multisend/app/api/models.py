import binascii
import uuid
import datetime

from app import db as models
from app import r as redis

from sqlalchemy.dialects.postgresql import JSON
from sqlalchemy.dialects.postgresql import UUID

from eth_keys import keys
from eth_utils import decode_hex
from app.api.encrypt import encrypt, decrypt


class MultiSendWatchTransaction(models.Model):
    """
    MultiSend transactions from block_chain
    """
    # id = models.Column(models.BigInteger, primary_key=True)
    id = models.Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)

    tx = models.Column(models.String(200), nullable=False, unique=True)
    currency = models.Column(models.String(16), nullable=True)
    block = models.Column(models.BigInteger, nullable=True)
    status = models.Column(models.String(16), default='un_confirm')

    info = models.Column(JSON)
    created_at = models.Column(models.DateTime, default=datetime.datetime.utcnow())
    updated_at = models.Column(models.DateTime, default=datetime.datetime.utcnow(), onupdate=datetime.datetime.utcnow())

    def __init__(self, tx, currency):
        self.tx = tx
        self.currency = currency

    def insert(self):
        models.session.add(self)
        models.session.commit()
        redis.hmset('pending', {self.tx: str(self.id)})

    @staticmethod
    def getAllByIDx(tx_id_list):
        return MultiSendWatchTransaction.query.filter(MultiSendWatchTransaction.id.in_(tx_id_list)).all()

    @staticmethod
    def getInfo(transaction):
        info = {}
        for i in ['blockHash', 'transactionHash', 'contractAddress']:
            if transaction[i]:
                info[i] = transaction[i].hex()
            else:
                info[i] = transaction[i]

        for i in ['blockNumber', 'cumulativeGasUsed', 'effectiveGasPrice', 'from', 'to',
                  'transactionIndex', 'status', 'type']:
            info[i] = transaction[i]

        return info

    @staticmethod
    def GetAllByTx(tx_list):
        return MultiSendWatchTransaction.objects.filter(tx__in=tx_list)

    @staticmethod
    def CreateTransactionFromChain(transaction, currency):
        status = 'valid'
        if transaction.status == 0:
            status = 'un_confirm'

        tx = MultiSendWatchTransaction(transaction.transactionHash.hex(), currency)
        tx.block = transaction.blockNumber
        tx.status = status
        tx.info = MultiSendWatchTransaction.getInfo(transaction)

        models.session.add(tx)
        models.session.commit()
        return

    @staticmethod
    def UpdateFromChain(transaction, currency):
        record = MultiSendWatchTransaction.query.filter_by(tx=transaction.transactionHash.hex()).first()
        # if None -> create
        if not record:
            return MultiSendWatchTransaction.CreateTransactionFromChain(transaction, currency)

        status = 'valid'
        if transaction.status == 0:
            status = 'invalid'

        record.block = transaction.blockNumber
        record.info = MultiSendWatchTransaction.getInfo(transaction)
        record.status = status
        models.session.commit()
        return


class MultiSendChain(models.Model):
    """
    MultiSend chain status record
    """
    id = models.Column(models.BigInteger, primary_key=True)
    status = models.Column(models.String(16), default='un_confirm')
    log = models.Column(JSON)
    data = models.Column(JSON)
    input = models.Column(JSON)
    parent = models.Column(models.BigInteger, nullable=True)

    created_at = models.Column(models.DateTime, default=datetime.datetime.utcnow())
    updated_at = models.Column(models.DateTime, default=datetime.datetime.utcnow(), onupdate=datetime.datetime.utcnow())

    def insert(self):
        models.session.add(self)
        models.session.commit()

    @staticmethod
    def GetUnconfirmed():
        return MultiSendChain.query.filter_by(status='un_confirm').all()


class PrivateAddress(models.Model):
    """
    Encrypted Private Key
    """
    id = models.Column(models.BigInteger, primary_key=True)
    public = models.Column(models.String(256), nullable=False, unique=True)
    private = models.Column(models.String(256), nullable=False, unique=True)

    created_at = models.Column(models.DateTime, default=datetime.datetime.utcnow())
    updated_at = models.Column(models.DateTime, default=datetime.datetime.utcnow(), onupdate=datetime.datetime.utcnow())


    @staticmethod
    def addPrivateKey(private):
        key = PrivateAddress()
        priv_key_bytes = decode_hex(private)
        priv_key = keys.PrivateKey(priv_key_bytes)
        pub_key = priv_key.public_key
        key.public = pub_key.to_checksum_address()

        key.private = encrypt(bytes(private, encoding='utf8'))

        if PrivateAddress.getPrivateKey(key.public) is None:
            models.session.add(key)
            models.session.commit()

        return key.public

    @staticmethod
    def getPrivateKey(public):
        record = PrivateAddress.query.filter_by(public=public).first()
        if record:
            return decrypt(record.private[:]).decode("utf-8")

        return record

    @staticmethod
    def _onlyPublic(record):
        return record.public

    @staticmethod
    def getAllPublicKey():
        record = PrivateAddress.query.all()
        return list(map(PrivateAddress._onlyPublic, record))
