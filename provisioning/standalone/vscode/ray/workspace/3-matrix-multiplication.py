"""
RAY MATRIX MULTIPLICATION EXAMPLE
---------------------------------
This script serves as a diagnostic tool to verify that:
1. The Ray client can connect to the KubeRay Head service.
2. The Ray Scheduler can successfully provision GPU resources.
3. The GPU worker has the correct drivers and PyTorch version installed.
4. Numerical computation (Matrix Multiplication) is performing at expected TFLOPS.

The script performs iterative matrix multiplications (AxB) on the GPU 
and calculates the aggregate performance in Teraflops.
"""

import ray
import time

# The internal cluster address for the Ray Head service.
RAY_HEAD_ADDRESS = "ray://raycluster-head-svc.workspace-kuberay-gpu.svc.cluster.local:10001"

# --- Configuration ---
NUM_TASKS   = 1      # Number of parallel Ray tasks to launch
MATRIX_SIZE = 10000  # Dimension of the square matrix (10k x 10k)
NUM_ITERS   = 1000   # Number of matrix multiplication iterations per task

def main():
    print(f"Connecting to Ray cluster at: {RAY_HEAD_ADDRESS}...")

    try:
        # runtime_env ensures the worker has the required packages.
        # If the image already has torch, Ray will skip installation if the version matches.
        runtime_env = {
            "pip": [
                "torch==2.4.1",
            ]
        }

        # Initialize the Ray connection with the defined environment.
        ray.init(
            address=RAY_HEAD_ADDRESS,
            runtime_env=runtime_env,
            ignore_reinit_error=True,
            log_to_driver=True,
        )
        print("Connected successfully to the Ray Cluster!\n")

        # The @ray.remote decorator defines a 'Task'.
        # num_gpus=1 ensures Ray schedules this only on a node with an available GPU.
        @ray.remote(num_gpus=1)
        def gpu_matrix_workload(task_id: int, matrix_size: int, num_iters: int) -> dict:
            import torch
            import time

            # Explicitly target the first available GPU allocated by Ray
            device = torch.device("cuda")
            gpu_name = torch.cuda.get_device_name(0)
            
            print(f"[Task {task_id}] Running on: {gpu_name}")
            print(f"[Task {task_id}] Matrix size: {matrix_size}x{matrix_size} | Iterations: {num_iters}")

            # Initialize random tensors directly on GPU memory (VRAM)
            # A 10k x 10k float32 matrix consumes ~400MB of VRAM.
            A = torch.randn(matrix_size, matrix_size, dtype=torch.float32, device=device)
            B = torch.randn(matrix_size, matrix_size, dtype=torch.float32, device=device)

            mem_used_gb = torch.cuda.memory_allocated() / 1e9
            print(f"[Task {task_id}] GPU memory allocated: {mem_used_gb:.2f} GB")

            # Warmup: Pre-heat the GPU kernels before starting the timer to avoid JIT overhead
            _ = torch.matmul(A, B)
            torch.cuda.synchronize() # Wait for GPU to finish the warmup pass

            # Start timing the heavy computation loop
            start = time.time()
            result = A.clone()

            for i in range(num_iters):
                # Core computation: Matrix Multiplication
                result = torch.matmul(result, B)

                # Periodic normalization to prevent numerical values from exploding (INF/NAN)
                if i % 10 == 0:
                    result = result / result.norm()

                # Progress log every 50 iterations
                if i % 50 == 0:
                    # Sync before measurement to get accurate timing of GPU progress
                    torch.cuda.synchronize() 
                    elapsed = time.time() - start
                    print(f"[Task {task_id}] Iter {i+1:03d}/{num_iters} | {elapsed:.1f}s elapsed")

            # Final synchronization to ensure total_time includes all GPU operations
            torch.cuda.synchronize()
            total_time = time.time() - start

            # Performance Calculation:
            # 2 * N^3 operations per matmul. 
            # We divide by 1e12 to get Teraflops (TFLOPS).
            checksum       = result.sum().item()
            flops_per_iter = 2 * (matrix_size ** 3)
            total_flops    = flops_per_iter * num_iters
            tflops         = total_flops / total_time / 1e12

            print(f"[Task {task_id}] ✅ Done | {total_time:.2f}s | {tflops:.2f} TFLOPS")

            # Release VRAM and clean up CUDA context
            torch.cuda.empty_cache()

            return {
                "task_id":     task_id,
                "device":      gpu_name,
                "total_time":  total_time,
                "tflops":      tflops,
                "checksum":    checksum,
            }

        # --- Execution Logic ---
        print(f"Submitting {NUM_TASKS} GPU tasks to the scheduler...")
        start_time = time.time()

        # Launch tasks asynchronously. This returns a list of 'Object References' (Futures).
        futures = [
            gpu_matrix_workload.remote(i, MATRIX_SIZE, NUM_ITERS)
            for i in range(NUM_TASKS)
        ]

        # ray.get() blocks until all tasks are finished and retrieves the results.
        results = ray.get(futures)
        elapsed = time.time() - start_time

        print("\n--- GPU Benchmark Results ---")
        total_tflops = 0
        for r in results:
            print(f"  Task {r['task_id']} | {r['device']} | {r['tflops']:.2f} TFLOPS")
            total_tflops += r['tflops']

        print(f"\nTotal wall time for all tasks : {elapsed:.2f}s")
        print(f"Aggregate Performance         : {total_tflops:.2f} TFLOPS")

    except Exception as e:
        print(f"\nCluster connection or execution error: {e}")
    finally:
        # Properly close the connection to the Head node and release resources
        if ray.is_initialized():
            ray.shutdown()
            print("Ray connection shut down.")

if __name__ == "__main__":
    main()