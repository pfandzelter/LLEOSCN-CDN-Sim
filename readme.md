# Large LEO Satellite Communication Network CDN PoP Simulator

Simulation tool for CDN replication in large low-earth orbit satellite access networks.

If you use this software in a publication, please cite it as:

### Text

T. Pfandzelter and D. Bermbach, **Edge (of the Earth) Replication: Optimizing Content Delivery in Large LEO Satellite Communication Networks**, in 21th IEEE/ACM International Symposium on Cluster, Cloud and Internet Computing (CCGrid 2021), May 2021.

### BibTeX

```bibtex
@inproceedings{pfandzelter2021edge,
    title     = "Edge (of the Earth) Replication: Optimizing Content Delivery in Large LEO Satellite Communication Networks",
    booktitle = "21th IEEE/ACM International Symposium on Cluster, Cloud and Internet Computing (CCGrid 2021)",
    author    = "Pfandzelter, Tobias and Bermbach, David",
    publisher = "IEEE",
    month     = may,
    year      = 2021
}
```

For a full list of publications, please see [our website](https://www.mcc.tu-berlin.de/menue/forschung/publikationen/parameter/en/).

## License

The code in this repository is licensed under the terms of the [MIT](./LICENSE) license.

All code in the `simulation` folder is based on the [SILLEO-SCNS project](https://github.com/Ben-Kempton/SILLEO-SCNS) and licensed under the [GNU General Public License Version 3](./simulation/LICENSE).

The [dataset of US cities](./data/us_cities.csv) is based on the [R `maps` package](https://github.com/adeckmyn/maps) and reproduced here for convenience.

The [dataset of Swiss cities](./data/swiss_cities.csv) is based on data from [OpenStreetMaps](https://openstreetmaps.org/) and reproduced here for convenience.

The [dataset of US cell towers](./data/us_cell_towers.csv) is based on public domain data by the [FCC](https://hifld-geoplatform.opendata.arcgis.com/datasets/cellular-towers) and reproduced here for convenience.

## Usage

### Installation (Amazon Linux 2)

`sh ./install.sh`

### Generate Workload

1. fill `workload.toml` (or choose one of the pre-configured workloads in the templates folder),
2. then run `sh ./workload.sh workload.toml`

### Run Simulation

`sh ./simulate.sh workload.toml`

You will find the results in the `results` sub-folder.

For performance reasons it is recommended to renice these processes to a niceness of -20 whereever possible, e.g. with `sudo renice -n -20 -p $(pgrep python3)`

### Calculate Caching

`sh ./caches.sh workload.toml`

### Run analysis

`sh ./analysis.sh workload.toml`
