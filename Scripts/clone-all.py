#!/usr/bin/env python3

import os
import subprocess as sp
import concurrent.futures
import argparse

def clone_all(base_repo, pkg, failed_pkgs):
    try:
        if os.path.exists(pkg):
            sp.run(["git", "-C", f"{pkg}", "pull"], check=True)
        else:
            sp.run(["git", "clone", f"https://{base_repo}{pkg}.git"], check=True)
            sp.run(["git", "-C", f"{pkg}", "remote", "set-url", "origin", f"https://{base_repo}{pkg}.git"])
            sp.run(["git", "-C", f"{pkg}", "remote", "set-url", "--push", "origin", f"git@{base_repo}{pkg}.git"])
    except Exception:
        failed_pkgs.append(pkg)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-j",
        type=int,
        default=1,
        choices=range(1, 31),
        help="Set the number of concurrent jobs e.g -j20"
    )
    args = parser.parse_args()

    pkgs_fpath = "common/packages"
    pkgs = []
    base_repo = "github.com/solus-packages/"

    try:
        with open(pkgs_fpath, "r") as file:
            for line in file:
                pkgs.append(line.strip())
    except FileNotFoundError:
        print(f"File not found: {pkgs_fpath}")

    failed_pkgs = []
    initial_loop = True

    try:
        while True:
            if initial_loop:
                with concurrent.futures.ThreadPoolExecutor(max_workers=args.j) as executor:
                    futures = {executor.submit(clone_all, base_repo, pkg, failed_pkgs) for pkg in pkgs}
                    initial_loop = False
            else:
                print(f"\n{failed_pkgs} packages failed")
                retry = input("Would you like to retry (y/n?)\n").lower()
                if retry not in ["yes", "y"]:
                    break
                with concurrent.futures.ThreadPoolExecutor(max_workers=args.j) as executor:
                    futures = {executor.submit(clone_all, base_repo, pkg, failed_pkgs) for pkg in failed_pkgs}
            if not failed_pkgs:
                break
    except KeyboardInterrupt:
        executor.shutdown(cancel_futures=True)
        print("Script terminated by ctrl+c.")

main()
