import random
from datetime import datetime, timedelta
import pymongo
from bson import ObjectId

# Database Connection
MONGO_URI = "mongodb://admin:supersecret@localhost:27017/?authSource=admin"
DB_NAME = "mqtt-streaming-server"
COLLECTION_NAME = "photos"

# Fake Data Sources
NAMES = ["Ion", "Maria", "Andrei", "Elena", "Radu", "Ana", "George", "Ioana", "Mihai", "Cristina", "Alexandru", "Gabriela", "Florin", "Daniela", "Vlad"]
SURNAMES = ["Popescu", "Ionescu", "Dumitru", "Stoica", "Radu", "Gheorghe", "Matei", "Florea", "Costea", "Marinescu", "Dinu", "Toma", "Stanciu", "Neagu", "Preda"]
JOBS = ["Inginer", "Programator", "Medic", "Profesor", "Contabil", "Sofer", "Manager", "Student", "Asistent", "Operator"]

def generate_random_photo():
    timestamp = datetime.now() - timedelta(days=random.randint(0, 30))
    nume = random.choice(SURNAMES)
    prenume = random.choice(NAMES)
    
    # Logic for Control Type (Mutual Exclusion usually, but boolean fields allow mix)
    # We will pick one main type and set it to True
    control_types = ["Angajare", "Periodic", "Adaptare", "Reluare", "Supraveghere", "Alte"]
    selected_control = random.choice(control_types)
    
    control_angajare = selected_control == "Angajare"
    control_periodic = selected_control == "Periodic"
    control_adaptare = selected_control == "Adaptare"
    control_reluare = selected_control == "Reluare"
    control_supraveghere = selected_control == "Supraveghere"
    control_alte = selected_control == "Alte"

    # Logic for Aviz (Opinion)
    aviz_types = ["APT", "APT Conditionat", "Inapt Temporar", "Inapt"]
    # Weight random choice towards APT (common)
    selected_aviz = random.choices(aviz_types, weights=[70, 15, 10, 5], k=1)[0]
    
    aviz_apt = selected_aviz == "APT"
    aviz_apt_conditionat = selected_aviz == "APT Conditionat"
    aviz_inapt_temporar = selected_aviz == "Inapt Temporar"
    aviz_inapt = selected_aviz == "Inapt"

    return {
        "timestamp": timestamp,
        "image_type": "jpeg",
        "device_id": f"device-{random.randint(1, 5)}",
        "text": f"Fake OCR for {nume} {prenume}",
        
        "unitate_medicala": "Clinica Test",
        "adresa_unitate_medicala": "Str. Testului nr 1",
        "telefon_unitate_medicala": "0700000000",
        "numar_fisa": f"FISA-{random.randint(1000, 9999)}",
        "societate_unitate": "Compania SRL",
        "adresa_angajator": "Bd. Muncii nr 10",
        "telefon_angajator": "0711111111",
        "nume": nume,
        "prenume": prenume,
        "cnp": f"{random.randint(1, 2)}{random.randint(50, 99)}{random.randint(10, 12)}{random.randint(10, 28)}123456",
        "profesie_functie": random.choice(JOBS),
        "loc_de_munca": "Bucuresti",
        
        "tip_control": f"Control {selected_control}", # Helper string field
        "control_angajare": control_angajare,
        "control_periodic": control_periodic,
        "control_adaptare": control_adaptare,
        "control_reluare": control_reluare,
        "control_supraveghere": control_supraveghere,
        "control_alte": control_alte,
        
        "aviz_medical": selected_aviz,
        "aviz_apt": aviz_apt,
        "aviz_apt_conditionat": aviz_apt_conditionat,
        "aviz_inapt_temporar": aviz_inapt_temporar,
        "aviz_inapt": aviz_inapt,
        
        "recomandari": "Nicio recomandare" if aviz_apt else "Reevaluare in 30 zile",
        "data": timestamp,
        "data_urm_examinari": timestamp + timedelta(days=365)
    }

def seed_data():
    try:
        client = pymongo.MongoClient(MONGO_URI)
        db = client[DB_NAME]
        collection = db[COLLECTION_NAME]
        
        current_count = collection.count_documents({})
        print(f"Current document count: {current_count}")
        
        records = [generate_random_photo() for _ in range(15)]
        
        result = collection.insert_many(records)
        print(f"Successfully inserted {len(result.inserted_ids)} records!")
        
        new_count = collection.count_documents({})
        print(f"New document count: {new_count}")
        
    except Exception as e:
        print(f"An error occurred: {e}")
        print("Ensure 'pymongo' is installed: pip install pymongo")

if __name__ == "__main__":
    seed_data()
