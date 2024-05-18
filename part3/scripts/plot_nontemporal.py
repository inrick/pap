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
    columns = ["Label", "Chunk size", "Max GB/s"]
    df2 = df[columns].pivot_table(
        index=["Label", "Chunk size"],
        values="Max GB/s",
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
        size_labels = [readable_bytes(n) for n in df3.index.values]
        max_bandwidth = df3.values
        ax.plot(max_bandwidth, label=f"{lbl}")

    if size_labels is not None:
        ax.grid(visible=True, linestyle="--", axis="both")
        ax.set_xticks(np.arange(len(size_labels)), size_labels, rotation=45, ha="center")
        # Resize plot to fit legend to the right of it.
        box = ax.get_position()
        ax.set_position([box.x0, box.y0, box.width - 0.12, box.height])
        ax.legend(
            loc="center left",
            bbox_to_anchor=(1, .5),
            ncol=1,
            fancybox=True,
            shadow=True,
        )
        ax.set_xlabel("Chunk size")
        ax.set_ylabel("GB/s")
        ax.set_xlim((0, len(size_labels)))
        ax.set_ylim(0)

    out_filename = filepath.with_suffix(".png").name
    out_file = PLOT_DIR / out_filename
    PLOT_DIR.mkdir(parents=True, exist_ok=True)
    fig.savefig(out_file)


if __name__ == "__main__":
    for f in RESULTS_DIR.iterdir():
        if f.name.startswith("nontemporal") and f.suffix == ".csv":
            plot_file(f)
