"""
A simple Python program to calculate the sum of numbers in an array.

Usage:
  1) Provide numbers as command-line arguments:
       python3 sum_array.py 1 2 3 4.5

  2) Or run without arguments and enter numbers when prompted:
       python3 sum_array.py
       Enter numbers separated by spaces: 1 2 3 4.5
"""

from __future__ import annotations
import sys
from typing import List

def convert_to_float_list(values: List[str]) -> List[float]:
    """Convert a list of strings to a list of floats with helpful errors."""
    numbers: List[float] = []
    for index, value in enumerate(values, start=1):
        try:
            numbers.append(float(value))
        except ValueError as error:
            raise ValueError(
                f"Item #{index} ('{value}') is not a number. "
                "Please provide only numeric values."
            ) from error
    return numbers


def calculate_sum(numbers: List[float]) -> float:
    """Return the sum of the given numbers list.

    This uses Python's built-in sum for clarity and performance.
    """
    return sum(numbers)


def main() -> None:
    # If numbers are provided as command-line args, use them; otherwise prompt.
    if len(sys.argv) > 1:
        raw_values = sys.argv[1:]
    else:
        user_input = input("Enter numbers separated by spaces: ").strip()
        raw_values = user_input.split() if user_input else []

    if not raw_values:
        print("No numbers provided. Nothing to sum.")
        return

    try:
        numbers = convert_to_float_list(raw_values)
    except ValueError as error:
        print(f"Error: {error}")
        sys.exit(1)

    total = calculate_sum(numbers)
    print(total)


if __name__ == "__main__":
    main()


