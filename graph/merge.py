# 
# This file is part of LLEOSCN-CDN-Sim (https://github.com/pfandzelter/LLEOSCN-CDN-Sim).
# Copyright (c) 2020 Tobias Pfandzelter.
# 
# This program is free software: you can redistribute it and/or modify  
# it under the terms of the GNU General Public License as published by  
# the Free Software Foundation, version 3.
#
# This program is distributed in the hope that it will be useful, but 
# WITHOUT ANY WARRANTY; without even the implied warranty of 
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU 
# General Public License for more details.
#
# You should have received a copy of the GNU General Public License 
# along with this program. If not, see <http://www.gnu.org/licenses/>.
#

import pandas
import os
import sys
import tqdm
import toml

if __name__ == "__main__":

    if len(sys.argv) != 2:
        exit(1)

    try:
        config = toml.load(sys.argv[1])
    except Exception as e:
        exit(e)

    results_folder_old = os.path.join(os.path.abspath(os.getcwd()), "workloads", config["name"], "data_old")
    results_folder_new = os.path.join(os.path.abspath(os.getcwd()), "workloads", config["name"], "data_new")

    results_folder_combined = os.path.join(os.path.abspath(os.getcwd()), "workloads", config["name"], "data")

    try:
        os.makedirs(results_folder_combined, exist_ok=True)
    except Exception as e:
        exit(e)

    for file in tqdm.tqdm(os.listdir(results_folder_old), desc="Merging Data..."):
        filename = os.fsdecode(file)
        if filename.endswith(".csv"):
            print(filename)
            data = pandas.read_csv(os.path.join(results_folder_old, filename), index_col="time")

            data_new = pandas.read_csv(os.path.join(results_folder_new, filename), index_col="time")

            data["SATELLITE-TIMEOUT"] = data_new["SATELLITE-TIMEOUT"]

            data["SATELLITE-VIRTUAL"] = data_new["SATELLITE-VIRTUAL"]

            data.to_csv(os.path.join(results_folder_combined, filename))
