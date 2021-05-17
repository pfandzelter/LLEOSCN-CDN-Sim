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

import multiprocessing as mp
import time
import pandas
from tqdm import tqdm, trange
import toml
import os
import sys

from simulation import Simulation

sys.path.append(os.path.abspath(os.getcwd()))
from loadgenerator.gen_load import single_workload

def simulate(steps, step_length, loc_file, result_file, load_file, city_file, request_amount):

    # constants
    # turning on animation is not recommended with more than 1 chunk
    ANIMATE = False

    EARTH_RADIUS = 6371000

    # Number of planes
    # PLANES = 24
    PLANES = 24

    # Number of nodes/plane
    # NODES = 66
    NODES = 66

    # Plane inclination (deg)
    INC = 53.0

    # Orbit Altitude (Km)
    ALTITUDE = 550

    # if true, will enable calculating network link-state
    MAKE_LINKS = True

    # Options: SPARSE, +GRID, IDEAL
    LINKING_METHOD = "+GRID"

    # Interval between calculations in ms
    FREQUENCY = 10

    # Cache strategy to use
    CACHE_STRATEGY = ["NONE", "GROUND-STATION", "SATELLITE", "SATELLITE-TIMEOUT", "VIRTUAL-POP"]

    s = Simulation(planes=int(PLANES), nodesPerPlane=int(NODES), inclination=float(INC), semiMajorAxis=float(ALTITUDE)*1000 + EARTH_RADIUS, timeStep=int(step_length), makeLinks=MAKE_LINKS, linkingMethod=LINKING_METHOD, captureImages=False, frequency=FREQUENCY, groundPtsFile=loc_file, animate=ANIMATE,enablePathCalc=True)

    for step in tqdm(steps, desc="simulating"):
        next_time = step*step_length
        workload = single_workload(load_file, city_file, request_amount, next_time)
        s.path_nodes = workload.loc[:, ["source", "origin", "size", "id"]].to_records(index=False)

        s.updateModel(next_time, result_file=result_file)
    
    if s.animation is not None:
        s.animation.terminate()

    time.sleep(2)

    s.terminate()


if __name__ == "__main__":
    try:
        config = toml.load(sys.argv[1])
    except Exception as e:
        exit(e)

    workload_folder = os.path.join(os.path.abspath(os.getcwd()), "workloads", config["name"])
    config_file = os.path.join(workload_folder, "config.toml")

    cfg = toml.load(config_file)

    steps = cfg["steps"]
    step_length = cfg["step_length"]

    load_file = os.path.join(workload_folder, cfg["loadfile"])

    loc_file = os.path.join(workload_folder, cfg["locations"])
    city_file = os.path.join(workload_folder, cfg["cities"])

    request_amount = cfg["requestamount"]

    result_file = os.path.join(workload_folder, "results", "r.csv")
    try:
        os.makedirs(os.path.join(workload_folder, "results"), exist_ok=True)
    except Exception as e:
        exit(e)
    
    def chunks(l, n):
        """Yield n number of striped chunks from l."""
        for i in range(0, n):
            yield l[i::n]

    p_list = []

    # leave one processor for other things
    CHUNK_C = mp.cpu_count() - 1

    for s in chunks(range(steps), CHUNK_C):
        kw = {
            "steps": s,
            "step_length": step_length,
            "loc_file": loc_file,
            "result_file": result_file,
            "load_file": load_file,
            "city_file": city_file,
            "request_amount": request_amount
        }

        p = mp.Process(target=simulate, kwargs=kw)
        p_list.append(p)

    for p in p_list:
        p.start()

    for p in p_list:
        p.join()
