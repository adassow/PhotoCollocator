from shutil import copyfile
import exifread
import os
import sys
import logging
from datetime import datetime

logger = logging.getLogger()
logger.setLevel(logging.INFO)
def get_crt_time(path):
    _, file_extension = os.path.splitext(path_name)
    if file_extension[1:].lower() not in ('mp4'):
        # logging.info("File {} not supported".format(path))
        return False

    f = open(path_name, 'rb')
    tags = exifread.process_file(f)
    st = os.stat(path)
    if not tags:
        logging.info("Exif create date not found for file {}".format(path))
        return datetime.fromtimestamp(st.st_ctime)
    create_time = tags.get('EXIF DateTimeOriginal')
    if not create_time:
        logging.info("Exif DateTimeOriginal not found for file {}. All data {}".format(path, tags.keys()))
        return datetime.fromtimestamp(st.st_ctime)
    return datetime.strptime(str(create_time), "%Y:%m:%d %H:%M:%S")

base_dir = sys.argv[1]
print(base_dir)
dest_dir = sys.argv[2]
for dirname, dirnames, filenames in os.walk(base_dir):
    for filename in filenames:
        path_name = os.path.join(dirname, filename)
        create_date = get_crt_time(path_name)
        if not create_date:
            continue


        directory = "{}/{}".format(dest_dir, create_date.strftime('%Y_%m'))
        _, file_extension = os.path.splitext(path_name)
        dest = "{}/{}".format(
            directory,
            create_date.strftime('%Y%m%d%H%M%S')+file_extension
        )
        if not os.path.exists(directory):
            os.makedirs(directory)
        logging.info("from:{} to:{}".format(path_name, dest))
        # print("from:{} to:{}".format(path_name, dest))
        #shutil.copyfile(path_name, dest)




