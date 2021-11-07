from app import db
import app.api.models

db.session.remove()
print(db.drop_all())
print(db.session.commit())
print(db.create_all())
print(db.session.commit())
