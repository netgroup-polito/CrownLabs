"""
RAY GPU SANITY CHECK
--------------------
Use this script to verify that your Ray cluster is properly 
allocating GPU resources and that PyTorch can communicate 
with the NVIDIA drivers.

It performs a basic hardware check and a quick matrix 
multiplication to ensure the GPU is fully functional.
"""

import ray

# The internal cluster address for the Ray Head service.
RAY_HEAD_ADDRESS = "ray://raycluster-head-svc.workspace-kuberay-gpu.svc.cluster.local:10001"

def main():
    print("Step 1: Connecting to the Ray Cluster...")
    
    try:
        # Initialize Ray. The runtime_env ensures PyTorch is available 
        # on the worker node before the task starts.
        ray.init(
            address=RAY_HEAD_ADDRESS,
            runtime_env={
                "pip": ["torch==2.4.1"]
            },
            ignore_reinit_error=True
        )
        print("✅ Connected successfully.\n")

        # The @ray.remote decorator turns this function into a Ray Task.
        # num_gpus=1 forces the scheduler to run this ONLY on a node with a GPU.
        @ray.remote(num_gpus=1)
        def verify_gpu_and_drivers() -> dict:
            import torch

            print("--- Running Hardware Diagnostics ---")
            
            # 1. Check if PyTorch can see the NVIDIA drivers
            cuda_available = torch.cuda.is_available()
            if not cuda_available:
                return {"status": "FAILED", "error": "CUDA is not available. Drivers may be missing."}

            # 2. Gather Hardware and Driver Info
            gpu_name = torch.cuda.get_device_name(0)
            cuda_version = torch.version.cuda
            torch_version = torch.__version__
            vram_gb = torch.cuda.get_device_properties(0).total_memory / 1e9

            print(f"GPU Model    : {gpu_name}")
            print(f"CUDA Version : {cuda_version}")
            print(f"PyTorch      : {torch_version}")
            print(f"Total VRAM   : {vram_gb:.1f} GB")

            # 3. Perform a quick Sanity Check (Simple Math on GPU)
            print("\n--- Running Math Sanity Check ---")
            try:
                # Create two random matrices directly on the GPU
                A = torch.randn(1000, 1000, device="cuda")
                B = torch.randn(1000, 1000, device="cuda")
                
                # Multiply them together
                _ = torch.matmul(A, B)
                
                # Force PyTorch to wait for the GPU to finish the calculation
                torch.cuda.synchronize()
                
                print("✅ Matrix multiplication successful!")
                math_success = True
                
            except Exception as e:
                print(f"❌ Matrix multiplication failed: {e}")
                math_success = False

            # Free up the GPU memory for the next user
            torch.cuda.empty_cache()

            return {
                "status": "SUCCESS" if math_success else "FAILED",
                "gpu_model": gpu_name,
                "cuda_version": cuda_version,
                "math_check_passed": math_success
            }

        # --- Execution Logic ---
        print("Step 2: Submitting GPU diagnostic task...")
        print("(If the cluster is scaling up a new GPU node, this may take a minute or two)\n")
        
        # .remote() submits the task, ray.get() waits for the result
        future = verify_gpu_and_drivers.remote()
        result = ray.get(future)

        # --- Final Output ---
        print("\n=== DIAGNOSTIC REPORT ===")
        print(f"Overall Status : {result['status']}")
        if result['status'] == "SUCCESS":
            print(f"Hardware       : {result['gpu_model']} (CUDA {result['cuda_version']})")
            print("Conclusion     : 🟢 The GPU worker is healthy and ready for workloads.")
        else:
            print(f"Error          : {result.get('error', 'Math operation failed')}")
            print("Conclusion     : 🔴 The GPU worker has an issue. Please contact your cluster admin.")

    except Exception as e:
        print(f"\n❌ Failed to execute test. Is the Ray cluster running? Error details: {e}")
        
    finally:
        # Always disconnect cleanly
        if ray.is_initialized():
            ray.shutdown()

if __name__ == "__main__":
    main()