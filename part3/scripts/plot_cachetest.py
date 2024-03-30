from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd


BASE_DIR = Path(__file__).resolve().parent.parent
RESULTS_DIR = BASE_DIR / "data" / "results"
PLOT_DIR = BASE_DIR / "data" / "plots"


def plot_file(filepath: Path):
    df = pd.read_csv(filepath)
    labels = df["Label"].unique()

    fig, ax = plt.subplots()
    figwidth = 8
    if len(labels) > 0:
        figwidth = 0.4 * len(df[df["Label"] == labels[0]])
    fig.set_size_inches(figwidth, 5)
    fig.tight_layout(pad=5)
    size_labels = None

    for lbl in labels:
        df2 = df[["Label", "Size label", "Chunk size", "Max GB/s"]][df["Label"] == lbl]
        size_labels = df2["Size label"].values
        max_bandwidth = df2["Max GB/s"].values
        ax.plot(max_bandwidth, label=lbl)

    if size_labels is not None:
        ax.grid(visible=True, linestyle="--", axis="both")
        ax.set_xticks(np.arange(len(size_labels)), size_labels, rotation=45, ha="center")
        ax.legend()
        ax.set_xlabel("Chunk size")
        ax.set_ylabel("GB/s")
        ax.set_ylim(0)

    out_filename = filepath.with_suffix(".png").name
    out_file = PLOT_DIR / out_filename
    PLOT_DIR.mkdir(parents=True, exist_ok=True)
    fig.savefig(out_file)


if __name__ == "__main__":
    for f in RESULTS_DIR.iterdir():
        if f.suffix == ".csv":
            plot_file(f)
