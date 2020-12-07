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