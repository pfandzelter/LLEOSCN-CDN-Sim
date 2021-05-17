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

import seaborn as sns
import pandas
import os
import sys
import tqdm
import toml

import matplotlib

matplotlib.rcParams['agg.path.chunksize'] = 10000

import matplotlib.pyplot as plt

if __name__ == "__main__":

    if len(sys.argv) != 2:
        exit(1)

    try:
        config = toml.load(sys.argv[1])
    except Exception as e:
        exit(e)

    results_folder = os.path.join(os.path.abspath(os.getcwd()), "workloads", config["name"], "data")

    for file in tqdm.tqdm(os.listdir(results_folder), desc="Generating Graphs..."):
        filename = os.fsdecode(file)
        if filename.endswith(".csv"):
            print(filename)
            # complete graph
            data = pandas.read_csv(os.path.join(results_folder, filename), index_col="time")

            fig, ax = plt.subplots(figsize=[15.0, 5.0])

            sns.lineplot(data=data, ax=ax)

            fig.savefig(os.path.join(results_folder, filename + ".png"), dpi=1000)

            ax.set_yscale("log")

            fig.savefig(os.path.join(results_folder, filename + "-log.png"), dpi=1000)

            plt.close(fig)

            # only ten minutes in the middle
            middle = 43200 #86400 / 2

            middle_data = data.iloc[middle-5*60:middle+5*60,]

            fig, ax = plt.subplots(figsize=[15.0, 5.0])

            sns.lineplot(data=middle_data, ax=ax)

            fig.savefig(os.path.join(results_folder, filename + "middle.png"), dpi=1000)

            ax.set_yscale("log")

            fig.savefig(os.path.join(results_folder, filename + "middle-log.png"), dpi=1000)

            plt.close(fig)

        else:
            continue