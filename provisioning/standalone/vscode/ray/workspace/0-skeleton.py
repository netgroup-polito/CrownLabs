"""
CODE SKELETON FOR A PROGRAM THAT USES RAY TO HANDLE DETACHED (GPU-BASED) JOBS
--------------------
Fill in the [INSERT HERE] sections to run your code on the cluster.
"""

import ray

# --- Your Remote Function ---
# Change num_gpus to num_cpus if your task doesn't need a graphics card
# Remember that, in case you are requesting a lot of resources ( e.g., N GPUs),
# the Ray cluster must have been configured to support your request.
# In case of errors, contact the cluster administrators.
@ray.remote(num_gpus=1)
def my_cluster_task():

    # [INSERT YOUR IMPORTS HERE] (e.g., import torch)

    print("Task is running on the cluster...")

    # [INSERT YOUR MODEL OR COMPUTE LOGIC HERE]

    # Return whatever data you need back on your local machine
    return {"status": "Done", "message": "My code ran successfully!"}


# --- Run the Cluster Job ---
def main():
    print("Connecting to Ray cluster...")

    # [INSERT ANY PIP PACKAGES YOU NEED HERE]
    ray.init(
        ignore_reinit_error=True
        # runtime_env={"pip": ["torch==2.4.1"]}
    )

    print("Submitting 1 task to the cluster...")
    # Submit the single task to the cluster
    # The '.remote()' tells Ray to run this task in a separate Worker, detached from the current program.
    future = my_cluster_task.remote()

    print("Waiting for cluster to finish...")
    # Fetch the result back to this script
    result = ray.get(future)

    print("\n--- Final Result ---")
    print(result)

    # Clean up
    ray.shutdown()
    print("Done!")

if __name__ == "__main__":
    main()