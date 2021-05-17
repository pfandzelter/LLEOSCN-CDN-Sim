# 
# This file is part of LLEOSCN-CDN-Sim (https://github.com/pfandzelter/LLEOSCN-CDN-Sim).
# Copyright (c) 2020 Ben S. Kempton, Tobias Pfandzelter.
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

# used to make program faster & responsive
import threading as td

# used to run animation in a different process
import multiprocessing as mp

# memory aligned arrays their manipulation for Python
import numpy as np

# custom classes
from constellation import Constellation

# Primarily using the write_gml() function...
import networkx as nx

# use to measure program performance (sim framerate)
import time

import os

from tqdm import tqdm, trange

# try to import numba funcs
try:
    import numba
    #import numba_funcs as nf
    USING_NUMBA = True
except ModuleNotFoundError:
    USING_NUMBA = False
    print("you probably do not have numba installed...")
    print("reverting to non-numba mode")


###############################################################################
#                               GLOBAL VARS                                   #
###############################################################################

EARTH_RADIUS = 6371000  # radius of Earth in meters

PNG_OUTPUT_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), "pics", "p")  # where to save images of the animation

MIN_SAT_ELEVATION = 30  # degrees

LANDMASS_OUTLINE_COLOR = (0.0, 0.0, 0.0)  # black, best contrast
EARTH_LAND_OPACITY = 1.0

EARTH_BASE_COLOR = (0.6, 0.6, 0.8)  # light blue, like water!
EARTH_OPACITY = 1.0

BACKGROUND_COLOR = (1.0, 1.0, 1.0)  # white

SAT_COLOR = (1.0, 0.0, 0.0)  # red, color of satellites
SAT_OPACITY = 1.0

GND_COLOR = (0.0, 1.0, 0.0)  # green, color of groundstations
GND_OPACITY = 1.0

ISL_LINK_COLOR = (0.9, 0.5, 0.1)  # yellow-brown, satellite-satellite links
ISL_LINK_OPACITY = 1.0
ISL_LINE_WIDTH = 3  # how wide to draw line in pixels

SGL_LINK_COLOR = (0.5, 0.9, 0.5)  # greenish? satellite-groundstation links
SGL_LINK_OPACITY = 0.75
SGL_LINE_WIDTH = 2  # how wide to draw line in pixels

PATH_LINK_COLOR = (0.8, 0.2, 0.8)  # purpleish? path links
PATH_LINK_OPACITY = 0.7
PATH_LINE_WIDTH = 13  # how wide to draw line in pixels

EARTH_SPHERE_POINTS = 5000  # higher = smoother earth model, slower to generate

SAT_POINT_SIZE = 9  # how big satellites are in (probably) screen pixels
GND_POINT_SIZE = 8  # how big ground points are in (probably) screen pixels

SECONDS_PER_DAY = 86400  # number of seconds per earth rotation (day)


###############################################################################
#                             SIMULATION CONTROL                              #
###############################################################################


class Simulation():

    def __init__(
            self,

            planes=1,
            nodesPerPlane=1,
            inclination=70,
            semiMajorAxis=6472000,
            timeStep=10,
            makeLinks=True,
            linkingMethod="SPARSE",
            animate=True,
            frequency=7,
            captureImages=False,
            captureInterpolation=1,
            groundPtsFile="city_data.txt",
            enablePathCalc=False,
            report_status=False):

        # constillation structure information
        self.num_planes = planes
        self.num_nodes_per_plane = nodesPerPlane
        self.plane_inclination = inclination
        self.semi_major_axis = semiMajorAxis
        self.min_communications_altitude = 100000
        self.min_sat_elevation = MIN_SAT_ELEVATION

        # path calculation
        self.path_nodes = []
        self.path_lengths = []
        self.path_links = []
        self.max_node_degree = -1

        # control flags
        self.animate = animate
        self.capture_images = captureImages
        self.capt_interpolation = 1
        self.make_links = makeLinks
        self.linking_method = linkingMethod  # used because it does not regenerate links
        self.enable_path_calculation = enablePathCalc
        self.report_status = report_status

        # timing control
        self.time_step = timeStep
        self.current_simulation_time = 0
        self.pause = False
        self.num_steps_to_run = -1
        self.frequency = frequency

        self.animation = None

        self.buckets_to_consider = set()
        self.last_step_nodes = set()
        self.sats_to_consider = set()

        # init the Constellation model
        self.model = Constellation(
            planes=self.num_planes,
            nodes_per_plane=self.num_nodes_per_plane,
            inclination=self.plane_inclination,
            semi_major_axis=self.semi_major_axis,
            minCommunicationsAltitude=self.min_communications_altitude,
            minSatElevation=self.min_sat_elevation,
            linkingMethod=self.linking_method)

        # add ground points to the constillation model
        # from the given file path
        # TODO:add error protection...
        data = []
        self.city_names = []
        self.city_names_gnd_id = {}
        gst_list = []
        with open(groundPtsFile, "r") as f:
            for line in f:
                my_line = []
                for word in line.split(","):
                    my_line.append(word)
                data.append(my_line)

        for i in range(1, len(data)):
            self.city_names.append(data[i][0])
            self.city_names_gnd_id[data[i][0]] = -1 * i
            id = self.model.addGroundPoint(float(data[i][1]), float(data[i][2]))
            gst_list.append([id, data[i][0]])
        

        # init the network design
        if self.make_links:
            self.initializeNetworkDesign()

        # so, after much effort it appears that I cannot control an
        # interactive vtk window externally. Therefore when running
        # with an animation, the animation class will have to drive
        # the simulation using an internal timer...
        if self.animate:

            from animation import Animation

            parent_conn, child_conn = mp.Pipe()

            kw = {
                "total_sats": self.model.total_sats,
                "sat_positions": self.model.getArrayOfSatPositions(),
                "ground_node_counter": -self.model.ground_node_counter,
                "gnd_positions": self.model.getArrayOfGndPositions(),
                "time_step": self.time_step,
                "current_simulation_time": self.current_simulation_time,
                "make_links": self.make_links,
                "capture_images": self.capture_images,
                "pipeConn": child_conn,
                "frequency": self.frequency
            }

            self.animation = mp.Process(target=Animation, kwargs=kw)
            self.animation.start()

            time.sleep(10)
            self.pipe_conn = parent_conn

    def terminate(self):
        if self.animate:
            self.animation.join()
            self.animation.close()        

    def initializeNetworkDesign(self):
        print("initalizing network design... ")
        self.max_isl_distance = self.model.calculateMaxISLDistance(
            self.min_communications_altitude)

        self.max_stg_distance = self.model.calculateMaxSpaceToGndDistance(
            self.min_sat_elevation)

        print("maxIsl: ", self.max_isl_distance)
        print("maxGtS: ", self.max_stg_distance)

        if self.linking_method == "+GRID":
            self.model.calculatePlusGridLinks(
                self.max_stg_distance,
                initialize=True,
                crosslink_interpolation=1)

        if self.linking_method == "SPARSE":
            self.model.calculatePlusGridLinks(
                self.max_stg_distance,
                initialize=True,
                crosslink_interpolation=self.model.total_sats + 1)

        print("done initalizing")

    def updateModel(self, new_time, result_file="results.csv"):
        """
        Update the model with a new time & recalculate links

        Function behaves differently depending on wether animate is true or not.
        If true, this func will be called from the updateAnimation() func
        If False, this will be called in a loop until some desired runtime is reached

        """

        time_1 = time.time()

        # grab initial time
        if self.num_steps_to_run > 0:
            self.num_steps_to_run -= 1
        elif self.num_steps_to_run == 0:
            self.pause = True
            self.num_steps_to_run = -1

        self.model.setConstillationTime(time=new_time)

        if self.make_links:
            if self.linking_method == "IDEAL":
                self.model.calculateIdealLinks(
                    self.max_isl_distance,
                    self.max_stg_distance)
            if self.linking_method == "+GRID":
                self.model.calculatePlusGridLinks(self.max_stg_distance, max_isl_range=self.max_isl_distance)
            if self.linking_method == "SPARSE":
                self.model.calculatePlusGridLinks(self.max_stg_distance)

        if self.animate:
            self.pipe_conn.send(["sat_positions", self.model.getArrayOfSatPositions()])
            self.pipe_conn.send(["gnd_positions", self.model.getArrayOfGndPositions()])
            self.pipe_conn.send(["links", self.model.getArrayOfLinks()])
            self.pipe_conn.send(["points",self.model.getArrayOfNodePositions()])
            self.pipe_conn.send(["total_sats", self.model.total_sats])
            self.pipe_conn.send(["enable_path_calculation",self.enable_path_calculation])
            self.pipe_conn.send(["pause", self.pause])
            self.pipe_conn.send(["current_simulation_time", new_time])

        if self.enable_path_calculation:

            nodes_to_consider = self.city_names_gnd_id.values()
            
            links = self.model.getGndToSatLinks(nodes_to_consider)

            sats_to_consider = set()

            for link in links:
                sats_to_consider.add(links[link][0])

            shortest_sat_paths = self.model.generateSatelliteShortestPaths(sats_to_consider) 

            paths = []
            
            for x in trange(len(self.path_nodes), desc="Path Calculations"):
                p = self.path_nodes[x]
                node_1 = p[0]
                node_2 = p[1]
                size = p[2]
                item_id = p[3]
                if (node_1 is not None) and (node_2 is not None):
                    id_1 = self.city_names_gnd_id[node_1]
                    id_2 = self.city_names_gnd_id[node_2]
                    sat_1_id = links[id_1][0]
                    
                    sat_2_id = links[id_2][0]

                    shortest_path = None
                    try:
                        if sat_1_id < sat_2_id:
                            shortest_path = [id_1] + shortest_sat_paths[sat_1_id][sat_2_id][0] + [id_2] 
                        else:
                            shortest_path = [id_2] + shortest_sat_paths[sat_2_id][sat_1_id][0] + [id_1] 
                            shortest_path.reverse()

                        p = []
                        if len(shortest_path) > 0:
                            for x in range(len(shortest_path)-1):
                                p.append([shortest_path[x], shortest_path[x+1]])
                        else:
                            p = None
                                            
                        paths.append([p, item_id, size])
                    except:
                        print(sat_1_id)
                        print(sat_2_id)

            with open(result_file + str(new_time) + "shortest_sat_paths", "w") as f:
                f.write("sat_1,sat_2,distance,path\n")
                for sat in range(len(shortest_sat_paths)):
                    for sat_2 in range(sat+1, len(shortest_sat_paths)):
                        path = shortest_sat_paths[sat][sat_2]
                        if len(path) <= 0:
                            continue
                        f.write(str(sat))
                        f.write(",")
                        f.write(str(sat_2))
                        f.write(",")
                        f.write(str(path[1])) # distance
                        f.write(",")
                        f.write(str(path[0][0]))
                        for n in range(1, len(path[0])):
                            f.write("|")
                            f.write(str(path[0][n]))
                        f.write("\n")
            
            with open(result_file + str(new_time) + "gnd_sat_links", "w") as f:
                f.write("gnd,sat,distance\n")
                for gnd_node in links:
                    f.write(str(gnd_node))
                    f.write(",")
                    f.write(str(links[gnd_node][0]))
                    f.write(",")
                    f.write(str(links[gnd_node][1]))
                    f.write("\n")

            with open(result_file + str(new_time) + "paths", "w") as f:
                f.write("item,bandwidth,path\n")
                for p in paths:
                    f.write(str(p[1]))
                    f.write(",")
                    f.write(str(p[2]))
                    f.write(",")
                    f.write(str(p[0][0][0]))
                    for n in p[0]:
                        f.write("|")
                        f.write(str(n[1]))
                    f.write("\n")

            if self.animate:
                self.path_links = []
                for path in tqdm(paths, desc="Record Updates"):
                    if path is not None:
                        self.path_links.append(path[0])
            
                self.pipe_conn.send(["path_links",self.path_links])

        self.current_simulation_time = new_time
        if self.animate:
                self.pipe_conn.send(["current_simulation_time", self.current_simulation_time])
        self.time_per_update = time.time() - time_1
            