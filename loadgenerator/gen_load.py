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

import numpy as np
import pandas
import os
from tqdm import trange, tqdm
import multiprocessing as mp
import toml
import sys

def single_workload(load_file, city_file, request_amount, time):
    load = pandas.read_csv(load_file)
    cities = pandas.read_csv(city_file)

    max_request_amount = request_amount

    time_request_amount = np.int64(np.floor(np.power(10, (-1 - ((-86400 + time) * time)/1399680000)) * max_request_amount ))

    request_amount = np.minimum(max_request_amount, time_request_amount)

    workload_data = {
        "source": np.random.default_rng(0).choice(cities["name"], p=cities["pop"]/np.sum(cities["pop"]), size=request_amount),
        "id": np.random.default_rng(0).choice(load["id"], p=load["pop"]/np.sum(load["pop"]), size=request_amount)
    }

    wl = pandas.DataFrame(data=workload_data, columns=["source", "id"])

    wl = wl.merge(load, how="left", on=["id"])
    
    return wl

def gen_load(base_path, workload):
    # load file
    # id (of item) | origin | size (in bytes)
    #
    # workload file
    # source (from cities) | id (id of item to retrieve) | time (in s)
    #

    try:
        os.makedirs(base_path, exist_ok=True)
    except Exception as e:
        exit(e)
    
    # read the list of cities as a pandas dataframee
    cities = pandas.read_csv(workload["cities"])

    # replace any weird spaces in city names
    cities.loc[:, "name"] = cities["name"].map(lambda x : str.replace(x, " ", "_"))

    # if we don't have population data, set population to 1 for every city
    if "pop" not in cities:
        cities["pop"] = 1

    # store city data in our workload folder as cities.csv
    cities[["name", "pop"]].to_csv(os.path.join(base_path, "cities.csv"), index=False)

    # read the origin locations
    origins = pandas.read_csv(workload["origins"])

    # build the locations file: all cities + all origins, needed for the simulation
    all_locations = cities[["name", "lat", "lon"]].append(origins[["name", "lat", "lon"]]).drop_duplicates(subset="name")

    all_locations.to_csv(os.path.join(base_path, "locations.csv"), index=False)

    # item popularity should follow a distribution where 50% of items are only accessed once, 40% accessed ~10 times, 5% accessed ~100 times, 5% accessed up to 10000 times
    # this is about an exponential distribution
    #
    # item size is 20% 1KB-3KB, about 40% 3-10KB, 30% 10-30KB, and 10% 30-300KB -> 20% 10^0 - 10^0.5, 40% 10^0.5 - 10^1, 30% 10^1 - 10^1.5, 10% 10^1.5 - 10^2.5
    # so size should be 10^normal distribution(mean=1.25, sd=0.5)
    items = pandas.DataFrame(data={"id": np.arange(workload["item_amount"]),
        "origin": np.random.default_rng(0).choice(origins["name"], p=origins["no_cache"]/np.sum(origins["no_cache"]), size=workload["item_amount"]),
        "pop": np.power(10., np.floor(np.random.default_rng(0).exponential(1., workload["item_amount"]))),
        "size": np.around(np.power(10., np.random.default_rng(0).normal(loc=1.25, scale=0.5,size=workload["item_amount"])) * 1000)
    })

    items.to_csv(os.path.join(base_path, "load.csv"), index=False)

    with open(os.path.join(base_path, "config.toml"), "w") as f:
        toml.dump({
            "step_length": workload["step_length"],
            "steps": workload["steps"],
            "requestamount": workload["request_amount"],
            "locations": "locations.csv",
            "cities": "cities.csv",
            "loadfile": "load.csv"
        }, f)

if __name__ == "__main__":
    try:
        config = toml.load(sys.argv[1])
    except Exception as e:
        exit(e)

    base_path = os.path.join(os.path.abspath(os.getcwd()), "workloads", config["name"])

    gen_load(base_path, config)