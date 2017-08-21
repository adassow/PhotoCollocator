import exifread
import os
import shutil
import sys

from datetime import datetime

base_dir = sys.argv[1]
print(base_dir)
dest_dir = sys.argv[2]
for dirname, dirnames, filenames in os.walk(base_dir):
    for filename in filenames:
        path_name = os.path.join(dirname, filename)
        f = open(path_name, 'rb')
        _, file_extension = os.path.splitext(path_name)
        # Return Exif tags
        tags = exifread.process_file(f)
        if file_extension[1:] == 'jpg':
            create_time = tags.get('EXIF DateTimeOriginal')
            if not create_time:
                print(filename)
                print(tags.keys())
                continue
            create_date = datetime.strptime(str(create_time), "%Y:%m:%d %H:%M:%S")
            directory = "{}/{}_{}".format(dest_dir, create_date.strftime('%Y'), create_date.strftime('%m'))
            dest = "{}/{}".format(
                directory,
                filename
            )
            if not os.path.exists(directory):
                os.makedirs(directory)
            print("from:{} to:{}".format(path_name, dest))
            #shutil.copyfile(path_name, dest)
