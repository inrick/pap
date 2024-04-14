from pathlib import Path

import matplotlib.pyplot as plt
import numpy as np
import pandas as pd


BASE_DIR = Path(__file__).resolve().parent.parent
RESULTS_DIR = BASE_DIR / "data" / "results"
PLOT_DIR = BASE_DIR / "data" / "plots"


def readable_bytes(n: int | float) -> str:
    units = ["B", "KB", "MB", "GB"]
    i = 0
    while i+1 < len(units) and 1024 <= n:
        n /= 1024
        i += 1
    return f"{n:.0f} {units[i]}"


def plot_file(filepath: Path):
    df = pd.read_csv(filepath)
    columns = ["Label", "Chunk size", "Offset size", "Max GB/s"]
    df2 = df[columns].pivot_table(
        index=["Label", "Offset size"],
        columns="Chunk size",
        values="Max GB/s",
        aggfunc="first",
    )

    labels = list(df2.index.get_level_values("Label").unique())

    fig, ax = plt.subplots()
    figwidth = 8
    if len(labels) > 0:
        figwidth = 0.5 * len(df2.loc[labels[0]])
    fig.set_size_inches(figwidth, 5)
    fig.tight_layout(pad=4)
    size_labels = None

    for lbl in labels:
        df3 = df2.loc[lbl]
        for chunksz, col in df3.items():
            size_labels = df3.reset_index()["Offset size"].values
            max_bandwidth = col.values
            ax.plot(max_bandwidth, label=f"{readable_bytes(chunksz)} ({lbl})")

    if size_labels is not None:
        size_labels = [readable_bytes(x) for x in size_labels]
        ax.grid(visible=True, linestyle="--", axis="both")
        ax.set_xticks(np.arange(len(size_labels)), size_labels, rotation=45, ha="center")
        # Resize plot to fit legend to the right of it.
        box = ax.get_position()
        ax.set_position([box.x0, box.y0, .82 * box.width, box.height])
        ax.legend(
            loc="center left",
            bbox_to_anchor=(1, .5),
            ncol=1,
            fancybox=True,
            shadow=True,
        )
        ax.set_xlabel("Offset size")
        ax.set_ylabel("GB/s")
        ax.set_ylim(0)

    out_filename = filepath.with_suffix(".png").name
    out_file = PLOT_DIR / out_filename
    PLOT_DIR.mkdir(parents=True, exist_ok=True)
    fig.savefig(out_file)


if __name__ == "__main__":
    for f in RESULTS_DIR.iterdir():
        if f.name.startswith("alignment") and f.suffix == ".csv":
            plot_file(f)
