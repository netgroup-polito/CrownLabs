"""
RAY CPU SUM TEST
----------------
Use this script to verify that your Ray cluster can successfully 
distribute tasks across standard CPU worker nodes. 

Instead of a heavy machine learning model, this script performs 
a parallelized mathematical operation (summing numbers) to ensure:
1. The Ray Head can communicate with CPU workers.
2. The scheduler properly chunks and distributes tasks.
3. Results can be gathered back to the driver without network drops.
"""

import ray
import time

# The internal cluster address for the Ray Head service.
RAY_HEAD_ADDRESS = "ray://raycluster-head-svc.workspace-kuberay-gpu.svc.cluster.local:10001"

# --- Configuration ---
N = 10_000_000_000               # The massive number we want to sum up to
NUM_CHUNKS = 2                   # Number of parallel tasks to split the work into
EXPECTED_SUM = N * (N + 1) // 2  # Mathematical truth (Gauss's formula) for validation

def main():
    print(f"Step 1: Connecting to Ray cluster at {RAY_HEAD_ADDRESS}...")
    
    try:
        # Initialize Ray. 
        # No runtime_env is needed here because this script only uses standard Python math.
        ray.init(
            address=RAY_HEAD_ADDRESS,
            ignore_reinit_error=True
        )
        print("✅ Connected successfully to the cluster.\n")

        # The @ray.remote decorator turns this function into a distributed Task.
        # num_cpus=1 explicitly tells Ray to run this on a CPU core.
        @ray.remote(num_cpus=1)
        def sum_cpu_chunk(start: int, end: int) -> int:
            import time
            
            # Print statements inside remote functions show up in the worker logs
            print(f"[CPU Worker] Computing sum for chunk: [{start:,} to {end:,})")
            
            # We use Gauss's formula to calculate the sum instantly. 
            # We add a small sleep to simulate computation time so we can actually 
            # observe the parallel execution in the cluster dashboard.
            time.sleep(1.0) 
            
            def gauss(k):
                return k * (k + 1) // 2
                
            return gauss(end - 1) - gauss(start - 1)

        # --- Execution Logic ---
        print(f"Step 2: Splitting the sum of 1 to {N:,} into {NUM_CHUNKS} parallel tasks...")
        start_time = time.time()

        # Calculate the size of each chunk
        chunk_size = N // NUM_CHUNKS
        futures = []

        # Loop to create and submit the chunks
        for i in range(NUM_CHUNKS):
            start = i * chunk_size + 1
            # Ensure the last chunk catches any remainder
            end = (i + 1) * chunk_size + 1 if i < NUM_CHUNKS - 1 else N + 1
            
            # .remote() submits the task to the cluster immediately
            futures.append(sum_cpu_chunk.remote(start, end))

        print(f"✅ Submitted {len(futures)} tasks to the CPU workers. Waiting for results...\n")

        # ray.get() blocks until all tasks are finished, returning a list of the partial sums
        partial_results = ray.get(futures)
        
        # Combine the distributed results locally
        total = sum(partial_results)
        elapsed = time.time() - start_time

        # --- Final Output ---
        print("=== CPU SUM TEST REPORT ===")
        print(f"Input Target   : {N:,}")
        print(f"Tasks Spawned  : {NUM_CHUNKS}")
        print(f"Expected Sum   : {EXPECTED_SUM:,}")
        print(f"Calculated Sum : {total:,}")
        print(f"Elapsed Time   : {elapsed:.2f} seconds\n")

        # Validation Check
        if total == EXPECTED_SUM:
            print("Conclusion     : 🟢 CPU scheduling and task execution are working perfectly.")
        else:
            delta = total - EXPECTED_SUM
            print(f"Conclusion     : 🔴 Mathematical mismatch (Delta: {delta}). Check worker stability.")

    except Exception as e:
        print(f"\n❌ Failed to execute test. Is the Ray cluster running? Error details: {e}")
        
    finally:
        # Always disconnect cleanly to free up cluster resources
        if ray.is_initialized():
            ray.shutdown()

if __name__ == "__main__":
    main()