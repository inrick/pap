import numpy as np

if __name__ == "__main__":
    trigargs = [-np.pi/2 + np.pi*n/8 for n in range(0,9)]
    trigargs.extend(-np.pi/2 + np.pi*np.random.rand(15))
    asinargs = np.random.rand(20)
    sqrtargs = np.random.rand(20)

    for x in [trigargs, asinargs, sqrtargs]:
        x.sort()

    for t in [
        ("Sin", np.sin, trigargs),
        ("Cos", np.cos, trigargs),
        ("Asin", np.asin, asinargs),
        ("Sqrt", np.sqrt, sqrtargs),
    ]:
        print(f"ref{t[0]} = []entry{{")
        for arg in t[2]:
            print(f"	{{{arg:.16f}, {t[1](arg):.16f}}},")
        print("}")
