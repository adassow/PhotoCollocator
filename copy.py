
from shutil import copyfile
import exifread
import os
import sys
import logging
import json
from datetime import datetime
import sqlite3
conn = sqlite3.connect('images.db')
conn.text_factory = str

logger = logging.getLogger()
logger.setLevel(logging.INFO)
def serialize_exif(exif):
    result = {}
    for k, v in exif.items():
        result[k] = repr(v)
    return json.dumps(result)
def get_file_data(path):
    ctime = None
    mtime = None
    ctime_exif = None
    exif_json = None
    _, file_extension = os.path.splitext(path_name)
    if file_extension[1:].lower() in ('jpg'):
        f = open(path_name, 'rb')
        exif = exifread.process_file(f)
        exif_json = serialize_exif(exif)
        create_time = exif.get('EXIF DateTimeOriginal')
        if create_time:
            ctime_exif = str(datetime.strptime(str(create_time), "%Y:%m:%d %H:%M:%S"))

    st = os.stat(path)
    ctime = str(datetime.fromtimestamp(st.st_ctime))
    mtime = str(datetime.fromtimestamp(st.st_mtime))

    return file_extension, ctime, mtime, exif_json, ctime_exif, st.st_size, None
def init_db(db):
    c = db.cursor()

    # Create table
    c.execute('''CREATE TABLE images
                 (id INTEGER PRIMARY KEY, 
                 file_name TEXT, 
                 ext TEXT,
                 crt_time TEXT, 
                 mod_time TEXT,
                 exif BLOB,
                 crt_time_exif TEXT,
                 size INTEGER, 
                 hash TEXT)''')
    # Save (commit) the changes
    db.commit()

def store(db, *args):
    c = db.cursor()
    logging.info(args[0])
    # Create tableh
    c.execute('''INSERT INTO images (
        file_name,ext,crt_time,mod_time,exif,crt_time_exif,size,hash) VALUES
        (?,?,?,?,?,?,?,?)''', args)

    db.commit()


base_dir = sys.argv[1]
print(base_dir)
init_db(conn)
for dirname, dirnames, filenames in os.walk(base_dir):
    for filename in filenames:
        path_name = os.path.join(dirname, filename)
        image_data = get_file_data(path_name)
        
        store(conn, path_name, *image_data)




conn.close()

# directory = "{}/{}".format(dest_dir, create_date.strftime('%Y_%m'))
# _, file_extension = os.path.splitext(path_name)
# dest = "{}/{}".format(
#     directory,
#     create_date.strftime('%Y%m%d%H%M%S')+file_extension
# )
# if not os.path.exists(directory):
#     os.makedirs(directory)
# shutil.copyfile(path_name, dest)