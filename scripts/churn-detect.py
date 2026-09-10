#!/usr/bin/env python3

# Script to detect non-deterministic SDK generation using `openapi-generation`

# Usage:
# ./churn-detect -s /path/to/openapi-spec.yaml -l python -n 100 -p /path/to/openapi-generation

"""
Flow:
1. Perform an initial generation of the SDK in a temporary directory.
2. Initialize a git repo in the SDK directory and commit the initial state.
3. Commence a loop that will run `n` times:
    - Perform a generation of the SDK
    - Compare the new state of the SDK with the previous state using `git diff`
      while taking care to ignore specific files that are bound to change.
      eg: package.json, pyproject.toml, *version.py, etc
    - If there are no changes, repeat steps 3 and 4
    - If there are changes, print the changes and exit
"""


import argparse
import difflib
import glob
import os
import shutil
import subprocess
import sys
import tempfile

VERBOSE = False


def verbose_log(message):
    if VERBOSE:
        print(message)


def parse_arguments():
    parser = argparse.ArgumentParser(
        description="Detect non-deterministic SDK generation using `openapi-generation`"
    )
    parser.add_argument(
        "-s", "--spec", required=True, help="Path to the OpenAPI specification file"
    )
    parser.add_argument(
        "-l",
        "--language",
        default="python",
        help="Programming language to generate the SDK in",
    )
    parser.add_argument(
        "-n", "--iterations", type=int, default=20, help="Number of iterations to run"
    )
    parser.add_argument(
        "-p", "--path", help="Path to the `openapi-generation` directory"
    )
    parser.add_argument("-o", "--output", help="Path to the output directory")
    parser.add_argument(
        "-e",
        "--exclude-patterns",
        nargs="+",
        default=["package.json", "pyproject.toml", "*version.py", ".speakeasy"],
        help="Patterns to exclude from the diff",
    )
    parser.add_argument("-v", "--verbose", action="store_true", help="Verbose output")
    args = parser.parse_args()
    if not args.path:
        # Find the `openapi-generation` directory in the user's home directory
        # by searching for the `cmd` directory
        args.path = glob.glob(
            os.path.join(os.path.expanduser("~"), "**", "openapi-generation"),
            recursive=True,
        )
        if not args.path:
            raise ValueError("Could not find `openapi-generation` directory")
        args.path = args.path[0]
    if not args.output:
        args.output = tempfile.mkdtemp(prefix=f"churn-detect-sdk-{args.language}-")
    global VERBOSE
    VERBOSE = args.verbose
    return args


def run_generation(args, iteration):
    cmd = [
        "go",
        "run",
        "-C",
        args.path,
        "cmd/generate/main.go",
        "-s",
        args.spec,
        "-o",
        args.output,
        "-l",
        args.language,
        "--license",
        "agpl-3.0-only",
        "--skip-compile",
    ]
    verbose_log(f"Running command: {' '.join(cmd)}")
    output_file = f"generation_{iteration}.log"
    try:
        with open(output_file, "w") as f:
            subprocess.run(
                cmd, check=True, stdout=f, stderr=subprocess.STDOUT, text=True
            )
    except subprocess.CalledProcessError as e:
        print(f"Failed to generate SDK. See {output_file} for details.")
        raise e


def call(cmd, **kwargs):
    try:
        subprocess.run(cmd, **kwargs)
    except subprocess.CalledProcessError as e:
        print(f"Command failed: {' '.join(cmd)}")
        print(e.stdout.decode() if e.stdout else str(e))
        print(e.stderr.decode() if e.stderr else str(e))
        raise e


def git_diff(args, iteration):
    filename = f"run_{iteration}"
    # Create git diff command with exclusions
    cmd = ["git", "diff"]
    for file in args.exclude_patterns:
        cmd.append(f":(exclude){file}")
    verbose_log(f"Running command: {' '.join(cmd)}")
    with open(f"{filename}.diff", "w") as f:
        subprocess.run(cmd, stdout=f, cwd=args.output)
    return f"{filename}.diff"


def run_and_check(args, iteration, diff_file="churn_diff.txt"):
    print(f"Running iteration {iteration}...")
    run_generation(args, iteration)
    after_state = git_diff(args, iteration)
    # Check for differences
    has_changes = False
    with open(after_state, "r") as after:
        after_content = after.readlines()
        verbose_log(f"After content: {after_content}")
        if after_content:
            has_changes = True
            shutil.copyfile(after_state, diff_file)

    if has_changes:
        print(f"Detected changes in iteration {iteration}!")
        print("Actual changes:\n")
        with open(diff_file, "r") as diff_file_handle:
            print(diff_file_handle.read())

    # Clean up temporary files
    os.remove(after_state)

    return not has_changes


def main():
    args = parse_arguments()
    print(f"Generating SDK in {args.output}")
    diff_file = "churn_diff.txt"
    if not run_and_check(args, 0, diff_file):
        # Initialize the git repo
        if not os.path.exists(os.path.join(args.output, ".git")):
            call(["git", "init"], cwd=args.output, check=True, capture_output=True)
        call(["git", "add", "."], cwd=args.output, check=True, capture_output=True)
        call(
            ["git", "commit", "-m", f"Initial commit"],
            cwd=args.output,
            check=True,
            capture_output=True,
        )
        verbose_log(f"Output repo initialized")
    # Run the generation `args.iterations` times
    for iteration in range(1, args.iterations + 1):
        if not run_and_check(args, iteration, diff_file):
            print(
                f"Non-deterministic generation detected after {iteration} iterations!"
            )
            print(f"See {diff_file} for details of the changes.")
            return 1

    # If we get here, no non-deterministic generation was detected
    print(
        f"Completed {args.iterations} iterations without detecting non-deterministic generation."
    )
    print("SDK generation appears to be deterministic.")

    return 0


if __name__ == "__main__":
    sys.exit(main())
