// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && (!cgo || purego)

package vk

// The Vulkan binding of vulkan.go without cgo: libvulkan is loaded at run
// time and every function is called through its pointer, with the
// structures of vulkan_types.go in Go memory. The code below is vulkan.go's,
// with its C types and helpers replaced. Keel patch.

import (
	"errors"
	"fmt"
	"image"
	"math"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// vkCall calls a Vulkan function; pointers to Go memory stay valid and in
// place for the call. Every argument is an integer, a handle or a pointer.
//
//go:uintptrescapes
func vkCall(fn uintptr, args ...uintptr) uintptr {
	r, _, _ := purego.SyscallN(fn, args...)
	return r
}

// intAndResult holds an integer and a result, as vulkan.go's helper does.
type intAndResult struct {
	uint uint32
	res  vkResult
}

// cString returns s NUL-terminated in Go memory; structures passed to
// Vulkan keep it alive for the call.
func cString(s string) *byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return &b[0]
}

// cFree is a no-op: cString memory is collected.
func cFree(unsafe.Pointer) {}

func goString(p *byte) string {
	if p == nil {
		return ""
	}
	n := 0
	for *(*byte)(unsafe.Add(unsafe.Pointer(p), n)) != 0 {
		n++
	}
	return string(unsafe.Slice(p, n))
}

func vkCreateInstance(f uintptr, pCreateInfo vkInstanceCreateInfo, pAllocator unsafe.Pointer, pInstance *vkInstance) vkResult {
	return vkResult(int32(vkCall(f, uintptr(unsafe.Pointer(&pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pInstance)))))
}

func vkDestroyInstance(f uintptr, instance vkInstance, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(instance), uintptr(pAllocator))
}

func vkEnumeratePhysicalDevices(f uintptr, instance vkInstance, pPhysicalDeviceCount *uint32, pPhysicalDevices *vkPhysicalDevice) vkResult {
	return vkResult(int32(vkCall(f, uintptr(instance), uintptr(unsafe.Pointer(pPhysicalDeviceCount)), uintptr(unsafe.Pointer(pPhysicalDevices)))))
}

func vkGetPhysicalDeviceQueueFamilyProperties(f uintptr, physicalDevice vkPhysicalDevice, pQueueFamilyPropertyCount *uint32, pQueueFamilyProperties *vkQueueFamilyProperties) {
	vkCall(f, uintptr(physicalDevice), uintptr(unsafe.Pointer(pQueueFamilyPropertyCount)), uintptr(unsafe.Pointer(pQueueFamilyProperties)))
}

func vkGetPhysicalDeviceFormatProperties(f uintptr, physicalDevice vkPhysicalDevice, format vkFormat, pFormatProperties *vkFormatProperties) {
	vkCall(f, uintptr(physicalDevice), uintptr(format), uintptr(unsafe.Pointer(pFormatProperties)))
}

func vkCreateDevice(f uintptr, physicalDevice vkPhysicalDevice, pCreateInfo vkDeviceCreateInfo, qinf vkDeviceQueueCreateInfo, pAllocator unsafe.Pointer, pDevice *vkDevice) vkResult {
	pCreateInfo.pQueueCreateInfos = &qinf
	return vkResult(int32(vkCall(f, uintptr(physicalDevice), uintptr(unsafe.Pointer(&pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pDevice)))))
}

func vkDestroyDevice(f uintptr, device vkDevice, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(pAllocator))
}

func vkGetDeviceQueue(f uintptr, device vkDevice, queueFamilyIndex uint32, queueIndex uint32, pQueue *vkQueue) {
	vkCall(f, uintptr(device), uintptr(queueFamilyIndex), uintptr(queueIndex), uintptr(unsafe.Pointer(pQueue)))
}

func vkCreateImageView(f uintptr, device vkDevice, pCreateInfo *vkImageViewCreateInfo, pAllocator unsafe.Pointer, pView *vkImageView) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pView)))))
}

func vkDestroyImageView(f uintptr, device vkDevice, imageView vkImageView, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(imageView), uintptr(pAllocator))
}

func vkCreateFramebuffer(f uintptr, device vkDevice, pCreateInfo vkFramebufferCreateInfo, pAllocator unsafe.Pointer, pFramebuffer *vkFramebuffer) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(&pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pFramebuffer)))))
}

func vkDestroyFramebuffer(f uintptr, device vkDevice, framebuffer vkFramebuffer, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(framebuffer), uintptr(pAllocator))
}

func vkDeviceWaitIdle(f uintptr, device vkDevice) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device))))
}

func vkQueueWaitIdle(f uintptr, queue vkQueue) vkResult {
	return vkResult(int32(vkCall(f, uintptr(queue))))
}

func vkCreateSemaphore(f uintptr, device vkDevice, pCreateInfo *vkSemaphoreCreateInfo, pAllocator unsafe.Pointer, pSemaphore *vkSemaphore) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pSemaphore)))))
}

func vkDestroySemaphore(f uintptr, device vkDevice, semaphore vkSemaphore, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(semaphore), uintptr(pAllocator))
}

func vkCreateRenderPass(f uintptr, device vkDevice, pCreateInfo vkRenderPassCreateInfo, subpassInf vkSubpassDescription, pAllocator unsafe.Pointer, pRenderPass *vkRenderPass) vkResult {
	pCreateInfo.pSubpasses = &subpassInf
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(&pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pRenderPass)))))
}

func vkDestroyRenderPass(f uintptr, device vkDevice, renderPass vkRenderPass, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(renderPass), uintptr(pAllocator))
}

func vkCreateCommandPool(f uintptr, device vkDevice, pCreateInfo *vkCommandPoolCreateInfo, pAllocator unsafe.Pointer, pCommandPool *vkCommandPool) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pCommandPool)))))
}

func vkDestroyCommandPool(f uintptr, device vkDevice, commandPool vkCommandPool, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(commandPool), uintptr(pAllocator))
}

func vkAllocateCommandBuffers(f uintptr, device vkDevice, pAllocateInfo *vkCommandBufferAllocateInfo, pCommandBuffers *vkCommandBuffer) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pAllocateInfo)), uintptr(unsafe.Pointer(pCommandBuffers)))))
}

func vkFreeCommandBuffers(f uintptr, device vkDevice, commandPool vkCommandPool, commandBufferCount uint32, pCommandBuffers *vkCommandBuffer) {
	vkCall(f, uintptr(device), uintptr(commandPool), uintptr(commandBufferCount), uintptr(unsafe.Pointer(pCommandBuffers)))
}

func vkBeginCommandBuffer(f uintptr, commandBuffer vkCommandBuffer, pBeginInfo vkCommandBufferBeginInfo) vkResult {
	return vkResult(int32(vkCall(f, uintptr(commandBuffer), uintptr(unsafe.Pointer(&pBeginInfo)))))
}

func vkEndCommandBuffer(f uintptr, commandBuffer vkCommandBuffer) vkResult {
	return vkResult(int32(vkCall(f, uintptr(commandBuffer))))
}

func vkQueueSubmit(f uintptr, queue vkQueue, pSubmits vkSubmitInfo, fence vkFence) vkResult {
	return vkResult(int32(vkCall(f, uintptr(queue), uintptr(1), uintptr(unsafe.Pointer(&pSubmits)), uintptr(fence))))
}

func vkCmdBeginRenderPass(f uintptr, commandBuffer vkCommandBuffer, pRenderPassBegin vkRenderPassBeginInfo, contents vkSubpassContents) {
	vkCall(f, uintptr(commandBuffer), uintptr(unsafe.Pointer(&pRenderPassBegin)), uintptr(contents))
}

func vkCmdEndRenderPass(f uintptr, commandBuffer vkCommandBuffer) {
	vkCall(f, uintptr(commandBuffer))
}

func vkCmdCopyBuffer(f uintptr, commandBuffer vkCommandBuffer, srcBuffer vkBuffer, dstBuffer vkBuffer, regionCount uint32, pRegions *vkBufferCopy) {
	vkCall(f, uintptr(commandBuffer), uintptr(srcBuffer), uintptr(dstBuffer), uintptr(regionCount), uintptr(unsafe.Pointer(pRegions)))
}

func vkCmdCopyBufferToImage(f uintptr, commandBuffer vkCommandBuffer, srcBuffer vkBuffer, dstImage vkImage, dstImageLayout vkImageLayout, regionCount uint32, pRegions *vkBufferImageCopy) {
	vkCall(f, uintptr(commandBuffer), uintptr(srcBuffer), uintptr(dstImage), uintptr(dstImageLayout), uintptr(regionCount), uintptr(unsafe.Pointer(pRegions)))
}

func vkCmdPipelineBarrier(f uintptr, commandBuffer vkCommandBuffer, srcStageMask vkPipelineStageFlags, dstStageMask vkPipelineStageFlags, dependencyFlags vkDependencyFlags, memoryBarrierCount uint32, pMemoryBarriers *vkMemoryBarrier, bufferMemoryBarrierCount uint32, pBufferMemoryBarriers *vkBufferMemoryBarrier, imageMemoryBarrierCount uint32, pImageMemoryBarriers *vkImageMemoryBarrier) {
	vkCall(f, uintptr(commandBuffer), uintptr(srcStageMask), uintptr(dstStageMask), uintptr(dependencyFlags), uintptr(memoryBarrierCount), uintptr(unsafe.Pointer(pMemoryBarriers)), uintptr(bufferMemoryBarrierCount), uintptr(unsafe.Pointer(pBufferMemoryBarriers)), uintptr(imageMemoryBarrierCount), uintptr(unsafe.Pointer(pImageMemoryBarriers)))
}

func vkCmdPushConstants(f uintptr, commandBuffer vkCommandBuffer, layout vkPipelineLayout, stageFlags vkShaderStageFlags, offset uint32, size uint32, pValues unsafe.Pointer) {
	vkCall(f, uintptr(commandBuffer), uintptr(layout), uintptr(stageFlags), uintptr(offset), uintptr(size), uintptr(pValues))
}

func vkCmdBindPipeline(f uintptr, commandBuffer vkCommandBuffer, pipelineBindPoint vkPipelineBindPoint, pipeline vkPipeline) {
	vkCall(f, uintptr(commandBuffer), uintptr(pipelineBindPoint), uintptr(pipeline))
}

func vkCmdBindVertexBuffers(f uintptr, commandBuffer vkCommandBuffer, firstBinding uint32, bindingCount uint32, pBuffers *vkBuffer, pOffsets *vkDeviceSize) {
	vkCall(f, uintptr(commandBuffer), uintptr(firstBinding), uintptr(bindingCount), uintptr(unsafe.Pointer(pBuffers)), uintptr(unsafe.Pointer(pOffsets)))
}

func vkCmdSetViewport(f uintptr, commandBuffer vkCommandBuffer, firstViewport uint32, viewportCount uint32, pViewports *vkViewport) {
	vkCall(f, uintptr(commandBuffer), uintptr(firstViewport), uintptr(viewportCount), uintptr(unsafe.Pointer(pViewports)))
}

func vkCmdBindIndexBuffer(f uintptr, commandBuffer vkCommandBuffer, buffer vkBuffer, offset vkDeviceSize, indexType vkIndexType) {
	vkCall(f, uintptr(commandBuffer), uintptr(buffer), uintptr(offset), uintptr(indexType))
}

func vkCmdDraw(f uintptr, commandBuffer vkCommandBuffer, vertexCount uint32, instanceCount uint32, firstVertex uint32, firstInstance uint32) {
	vkCall(f, uintptr(commandBuffer), uintptr(vertexCount), uintptr(instanceCount), uintptr(firstVertex), uintptr(firstInstance))
}

func vkCmdDrawIndexed(f uintptr, commandBuffer vkCommandBuffer, indexCount uint32, instanceCount uint32, firstIndex uint32, vertexOffset int32, firstInstance uint32) {
	vkCall(f, uintptr(commandBuffer), uintptr(indexCount), uintptr(instanceCount), uintptr(firstIndex), uintptr(vertexOffset), uintptr(firstInstance))
}

func vkCmdBindDescriptorSets(f uintptr, commandBuffer vkCommandBuffer, pipelineBindPoint vkPipelineBindPoint, layout vkPipelineLayout, firstSet uint32, descriptorSetCount uint32, pDescriptorSets *vkDescriptorSet, dynamicOffsetCount uint32, pDynamicOffsets *uint32) {
	vkCall(f, uintptr(commandBuffer), uintptr(pipelineBindPoint), uintptr(layout), uintptr(firstSet), uintptr(descriptorSetCount), uintptr(unsafe.Pointer(pDescriptorSets)), uintptr(dynamicOffsetCount), uintptr(unsafe.Pointer(pDynamicOffsets)))
}

func vkCmdCopyImageToBuffer(f uintptr, commandBuffer vkCommandBuffer, srcImage vkImage, srcImageLayout vkImageLayout, dstBuffer vkBuffer, regionCount uint32, pRegions *vkBufferImageCopy) {
	vkCall(f, uintptr(commandBuffer), uintptr(srcImage), uintptr(srcImageLayout), uintptr(dstBuffer), uintptr(regionCount), uintptr(unsafe.Pointer(pRegions)))
}

func vkCmdDispatch(f uintptr, commandBuffer vkCommandBuffer, groupCountX uint32, groupCountY uint32, groupCountZ uint32) {
	vkCall(f, uintptr(commandBuffer), uintptr(groupCountX), uintptr(groupCountY), uintptr(groupCountZ))
}

func vkCreateImage(f uintptr, device vkDevice, pCreateInfo *vkImageCreateInfo, pAllocator unsafe.Pointer, pImage *vkImage) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pImage)))))
}

func vkDestroyImage(f uintptr, device vkDevice, image vkImage, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(image), uintptr(pAllocator))
}

func vkGetImageMemoryRequirements(f uintptr, device vkDevice, image vkImage, pMemoryRequirements *vkMemoryRequirements) {
	vkCall(f, uintptr(device), uintptr(image), uintptr(unsafe.Pointer(pMemoryRequirements)))
}

func vkAllocateMemory(f uintptr, device vkDevice, pAllocateInfo *vkMemoryAllocateInfo, pAllocator unsafe.Pointer, pMemory *vkDeviceMemory) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pAllocateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pMemory)))))
}

func vkBindImageMemory(f uintptr, device vkDevice, image vkImage, memory vkDeviceMemory, memoryOffset vkDeviceSize) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(image), uintptr(memory), uintptr(memoryOffset))))
}

func vkFreeMemory(f uintptr, device vkDevice, memory vkDeviceMemory, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(memory), uintptr(pAllocator))
}

func vkGetPhysicalDeviceMemoryProperties(f uintptr, physicalDevice vkPhysicalDevice, pMemoryProperties *vkPhysicalDeviceMemoryProperties) {
	vkCall(f, uintptr(physicalDevice), uintptr(unsafe.Pointer(pMemoryProperties)))
}

func vkCreateSampler(f uintptr, device vkDevice, pCreateInfo *vkSamplerCreateInfo, pAllocator unsafe.Pointer, pSampler *vkSampler) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pSampler)))))
}

func vkDestroySampler(f uintptr, device vkDevice, sampler vkSampler, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(sampler), uintptr(pAllocator))
}

func vkCreateBuffer(f uintptr, device vkDevice, pCreateInfo *vkBufferCreateInfo, pAllocator unsafe.Pointer, pBuffer *vkBuffer) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pBuffer)))))
}

func vkDestroyBuffer(f uintptr, device vkDevice, buffer vkBuffer, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(buffer), uintptr(pAllocator))
}

func vkGetBufferMemoryRequirements(f uintptr, device vkDevice, buffer vkBuffer, pMemoryRequirements *vkMemoryRequirements) {
	vkCall(f, uintptr(device), uintptr(buffer), uintptr(unsafe.Pointer(pMemoryRequirements)))
}

func vkBindBufferMemory(f uintptr, device vkDevice, buffer vkBuffer, memory vkDeviceMemory, memoryOffset vkDeviceSize) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(buffer), uintptr(memory), uintptr(memoryOffset))))
}

func vkCreateShaderModule(f uintptr, device vkDevice, pCreateInfo vkShaderModuleCreateInfo, pAllocator unsafe.Pointer, pShaderModule *vkShaderModule) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(&pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pShaderModule)))))
}

func vkDestroyShaderModule(f uintptr, device vkDevice, shaderModule vkShaderModule, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(shaderModule), uintptr(pAllocator))
}

func vkCreateGraphicsPipelines(f uintptr, device vkDevice, pipelineCache vkPipelineCache, pCreateInfo vkGraphicsPipelineCreateInfo, dynInf vkPipelineDynamicStateCreateInfo, blendInf vkPipelineColorBlendStateCreateInfo, vertexInf vkPipelineVertexInputStateCreateInfo, viewportInf vkPipelineViewportStateCreateInfo, pAllocator unsafe.Pointer, pPipelines *vkPipeline) vkResult {
	pCreateInfo.pDynamicState = &dynInf
	pCreateInfo.pViewportState = &viewportInf
	pCreateInfo.pColorBlendState = &blendInf
	pCreateInfo.pVertexInputState = &vertexInf
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(pipelineCache), uintptr(1), uintptr(unsafe.Pointer(&pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pPipelines)))))
}

func vkDestroyPipeline(f uintptr, device vkDevice, pipeline vkPipeline, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(pipeline), uintptr(pAllocator))
}

func vkCreatePipelineLayout(f uintptr, device vkDevice, pCreateInfo vkPipelineLayoutCreateInfo, pAllocator unsafe.Pointer, pPipelineLayout *vkPipelineLayout) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(&pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pPipelineLayout)))))
}

func vkDestroyPipelineLayout(f uintptr, device vkDevice, pipelineLayout vkPipelineLayout, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(pipelineLayout), uintptr(pAllocator))
}

func vkCreateDescriptorSetLayout(f uintptr, device vkDevice, pCreateInfo vkDescriptorSetLayoutCreateInfo, pAllocator unsafe.Pointer, pSetLayout *vkDescriptorSetLayout) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(&pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pSetLayout)))))
}

func vkDestroyDescriptorSetLayout(f uintptr, device vkDevice, descriptorSetLayout vkDescriptorSetLayout, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(descriptorSetLayout), uintptr(pAllocator))
}

func vkMapMemory(f uintptr, device vkDevice, memory vkDeviceMemory, offset vkDeviceSize, size vkDeviceSize, flags vkMemoryMapFlags, ppData *unsafe.Pointer) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(memory), uintptr(offset), uintptr(size), uintptr(flags), uintptr(unsafe.Pointer(ppData)))))
}

func vkUnmapMemory(f uintptr, device vkDevice, memory vkDeviceMemory) {
	vkCall(f, uintptr(device), uintptr(memory))
}

func vkResetCommandBuffer(f uintptr, commandBuffer vkCommandBuffer, flags vkCommandBufferResetFlags) vkResult {
	return vkResult(int32(vkCall(f, uintptr(commandBuffer), uintptr(flags))))
}

func vkCreateDescriptorPool(f uintptr, device vkDevice, pCreateInfo vkDescriptorPoolCreateInfo, pAllocator unsafe.Pointer, pDescriptorPool *vkDescriptorPool) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(&pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pDescriptorPool)))))
}

func vkDestroyDescriptorPool(f uintptr, device vkDevice, descriptorPool vkDescriptorPool, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(descriptorPool), uintptr(pAllocator))
}

func vkAllocateDescriptorSets(f uintptr, device vkDevice, pAllocateInfo vkDescriptorSetAllocateInfo, pDescriptorSets *vkDescriptorSet) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(&pAllocateInfo)), uintptr(unsafe.Pointer(pDescriptorSets)))))
}

func vkFreeDescriptorSets(f uintptr, device vkDevice, descriptorPool vkDescriptorPool, descriptorSetCount uint32, pDescriptorSets *vkDescriptorSet) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(descriptorPool), uintptr(descriptorSetCount), uintptr(unsafe.Pointer(pDescriptorSets)))))
}

func vkUpdateDescriptorSets(f uintptr, device vkDevice, pDescriptorWrite vkWriteDescriptorSet, descriptorCopyCount uint32, pDescriptorCopies unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(1), uintptr(unsafe.Pointer(&pDescriptorWrite)), uintptr(descriptorCopyCount), uintptr(pDescriptorCopies))
}

func vkResetDescriptorPool(f uintptr, device vkDevice, descriptorPool vkDescriptorPool, flags vkDescriptorPoolResetFlags) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(descriptorPool), uintptr(flags))))
}

func vkCmdBlitImage(f uintptr, commandBuffer vkCommandBuffer, srcImage vkImage, srcImageLayout vkImageLayout, dstImage vkImage, dstImageLayout vkImageLayout, regionCount uint32, pRegions *vkImageBlit, filter vkFilter) {
	vkCall(f, uintptr(commandBuffer), uintptr(srcImage), uintptr(srcImageLayout), uintptr(dstImage), uintptr(dstImageLayout), uintptr(regionCount), uintptr(unsafe.Pointer(pRegions)), uintptr(filter))
}

func vkCmdCopyImage(f uintptr, commandBuffer vkCommandBuffer, srcImage vkImage, srcImageLayout vkImageLayout, dstImage vkImage, dstImageLayout vkImageLayout, regionCount uint32, pRegions *vkImageCopy) {
	vkCall(f, uintptr(commandBuffer), uintptr(srcImage), uintptr(srcImageLayout), uintptr(dstImage), uintptr(dstImageLayout), uintptr(regionCount), uintptr(unsafe.Pointer(pRegions)))
}

func vkCreateComputePipelines(f uintptr, device vkDevice, pipelineCache vkPipelineCache, createInfoCount uint32, pCreateInfos *vkComputePipelineCreateInfo, pAllocator unsafe.Pointer, pPipelines *vkPipeline) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(pipelineCache), uintptr(createInfoCount), uintptr(unsafe.Pointer(pCreateInfos)), uintptr(pAllocator), uintptr(unsafe.Pointer(pPipelines)))))
}

func vkCreateFence(f uintptr, device vkDevice, pCreateInfo *vkFenceCreateInfo, pAllocator unsafe.Pointer, pFence *vkFence) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pFence)))))
}

func vkDestroyFence(f uintptr, device vkDevice, fence vkFence, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(fence), uintptr(pAllocator))
}

func vkWaitForFences(f uintptr, device vkDevice, fenceCount uint32, pFences *vkFence, waitAll vkBool32, timeout uint64) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(fenceCount), uintptr(unsafe.Pointer(pFences)), uintptr(waitAll), uintptr(timeout))))
}

func vkResetFences(f uintptr, device vkDevice, fenceCount uint32, pFences *vkFence) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(fenceCount), uintptr(unsafe.Pointer(pFences)))))
}

func vkGetPhysicalDeviceProperties(f uintptr, physicalDevice vkPhysicalDevice, pProperties *vkPhysicalDeviceProperties) {
	vkCall(f, uintptr(physicalDevice), uintptr(unsafe.Pointer(pProperties)))
}

func vkGetPhysicalDeviceSurfaceSupportKHR(f uintptr, physicalDevice vkPhysicalDevice, queueFamilyIndex uint32, surface vkSurfaceKHR, pSupported *vkBool32) vkResult {
	return vkResult(int32(vkCall(f, uintptr(physicalDevice), uintptr(queueFamilyIndex), uintptr(surface), uintptr(unsafe.Pointer(pSupported)))))
}

func vkDestroySurfaceKHR(f uintptr, instance vkInstance, surface vkSurfaceKHR, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(instance), uintptr(surface), uintptr(pAllocator))
}

func vkGetPhysicalDeviceSurfaceFormatsKHR(f uintptr, physicalDevice vkPhysicalDevice, surface vkSurfaceKHR, pSurfaceFormatCount *uint32, pSurfaceFormats *vkSurfaceFormatKHR) vkResult {
	return vkResult(int32(vkCall(f, uintptr(physicalDevice), uintptr(surface), uintptr(unsafe.Pointer(pSurfaceFormatCount)), uintptr(unsafe.Pointer(pSurfaceFormats)))))
}

func vkGetPhysicalDeviceSurfacePresentModesKHR(f uintptr, physicalDevice vkPhysicalDevice, surface vkSurfaceKHR, pPresentModeCount *uint32, pPresentModes *vkPresentModeKHR) vkResult {
	return vkResult(int32(vkCall(f, uintptr(physicalDevice), uintptr(surface), uintptr(unsafe.Pointer(pPresentModeCount)), uintptr(unsafe.Pointer(pPresentModes)))))
}

func vkGetPhysicalDeviceSurfaceCapabilitiesKHR(f uintptr, physicalDevice vkPhysicalDevice, surface vkSurfaceKHR, pSurfaceCapabilities *vkSurfaceCapabilitiesKHR) vkResult {
	return vkResult(int32(vkCall(f, uintptr(physicalDevice), uintptr(surface), uintptr(unsafe.Pointer(pSurfaceCapabilities)))))
}

func vkCreateSwapchainKHR(f uintptr, device vkDevice, pCreateInfo *vkSwapchainCreateInfoKHR, pAllocator unsafe.Pointer, pSwapchain *vkSwapchainKHR) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pSwapchain)))))
}

func vkDestroySwapchainKHR(f uintptr, device vkDevice, swapchain vkSwapchainKHR, pAllocator unsafe.Pointer) {
	vkCall(f, uintptr(device), uintptr(swapchain), uintptr(pAllocator))
}

func vkGetSwapchainImagesKHR(f uintptr, device vkDevice, swapchain vkSwapchainKHR, pSwapchainImageCount *uint32, pSwapchainImages *vkImage) vkResult {
	return vkResult(int32(vkCall(f, uintptr(device), uintptr(swapchain), uintptr(unsafe.Pointer(pSwapchainImageCount)), uintptr(unsafe.Pointer(pSwapchainImages)))))
}

func vkAcquireNextImageKHR(f uintptr, device vkDevice, swapchain vkSwapchainKHR, timeout uint64, semaphore vkSemaphore, fence vkFence) intAndResult {
	var res intAndResult
	res.res = vkResult(int32(vkCall(f, uintptr(device), uintptr(swapchain), uintptr(timeout), uintptr(semaphore), uintptr(fence), uintptr(unsafe.Pointer(&res.uint)))))
	return res
}

func vkQueuePresentKHR(f uintptr, queue vkQueue, pPresentInfo vkPresentInfoKHR) vkResult {
	return vkResult(int32(vkCall(f, uintptr(queue), uintptr(unsafe.Pointer(&pPresentInfo)))))
}

type (
	AttachmentLoadOp      = vkAttachmentLoadOp
	AccessFlags           = vkAccessFlags
	BlendFactor           = vkBlendFactor
	Buffer                = vkBuffer
	BufferImageCopy       = vkBufferImageCopy
	BufferMemoryBarrier   = vkBufferMemoryBarrier
	BufferUsageFlags      = vkBufferUsageFlags
	CommandPool           = vkCommandPool
	CommandBuffer         = vkCommandBuffer
	DependencyFlags       = vkDependencyFlags
	DescriptorPool        = vkDescriptorPool
	DescriptorPoolSize    = vkDescriptorPoolSize
	DescriptorSet         = vkDescriptorSet
	DescriptorSetLayout   = vkDescriptorSetLayout
	DescriptorType        = vkDescriptorType
	Device                = vkDevice
	DeviceMemory          = vkDeviceMemory
	DeviceSize            = vkDeviceSize
	Fence                 = vkFence
	Queue                 = vkQueue
	IndexType             = vkIndexType
	Image                 = vkImage
	ImageBlit             = vkImageBlit
	ImageCopy             = vkImageCopy
	ImageLayout           = vkImageLayout
	ImageMemoryBarrier    = vkImageMemoryBarrier
	ImageUsageFlags       = vkImageUsageFlags
	ImageView             = vkImageView
	Instance              = vkInstance
	Filter                = vkFilter
	Format                = vkFormat
	FormatFeatureFlags    = vkFormatFeatureFlags
	Framebuffer           = vkFramebuffer
	MemoryBarrier         = vkMemoryBarrier
	MemoryPropertyFlags   = vkMemoryPropertyFlags
	Pipeline              = vkPipeline
	PipelineBindPoint     = vkPipelineBindPoint
	PipelineLayout        = vkPipelineLayout
	PipelineStageFlags    = vkPipelineStageFlags
	PhysicalDevice        = vkPhysicalDevice
	PrimitiveTopology     = vkPrimitiveTopology
	PushConstantRange     = vkPushConstantRange
	QueueFamilyProperties = vkQueueFamilyProperties
	QueueFlags            = vkQueueFlags
	RenderPass            = vkRenderPass
	Sampler               = vkSampler
	SamplerMipmapMode     = vkSamplerMipmapMode
	Semaphore             = vkSemaphore
	ShaderModule          = vkShaderModule
	ShaderStageFlags      = vkShaderStageFlags
	SubpassDependency     = vkSubpassDependency
	Viewport              = vkViewport
	WriteDescriptorSet    = vkWriteDescriptorSet

	Surface             = vkSurfaceKHR
	SurfaceCapabilities = vkSurfaceCapabilitiesKHR

	Swapchain = vkSwapchainKHR
)

type VertexInputBindingDescription struct {
	Binding int
	Stride  int
}

type VertexInputAttributeDescription struct {
	Location int
	Binding  int
	Format   Format
	Offset   int
}

type DescriptorSetLayoutBinding struct {
	Binding        int
	DescriptorType DescriptorType
	StageFlags     ShaderStageFlags
}

type Error int32

const (
	FORMAT_R8G8B8A8_UNORM      Format = _VK_FORMAT_R8G8B8A8_UNORM
	FORMAT_B8G8R8A8_SRGB       Format = _VK_FORMAT_B8G8R8A8_SRGB
	FORMAT_R8G8B8A8_SRGB       Format = _VK_FORMAT_R8G8B8A8_SRGB
	FORMAT_R16_SFLOAT          Format = _VK_FORMAT_R16_SFLOAT
	FORMAT_R32_SFLOAT          Format = _VK_FORMAT_R32_SFLOAT
	FORMAT_R32G32_SFLOAT       Format = _VK_FORMAT_R32G32_SFLOAT
	FORMAT_R32G32B32_SFLOAT    Format = _VK_FORMAT_R32G32B32_SFLOAT
	FORMAT_R32G32B32A32_SFLOAT Format = _VK_FORMAT_R32G32B32A32_SFLOAT

	FORMAT_FEATURE_COLOR_ATTACHMENT_BIT            FormatFeatureFlags = _VK_FORMAT_FEATURE_COLOR_ATTACHMENT_BIT
	FORMAT_FEATURE_COLOR_ATTACHMENT_BLEND_BIT      FormatFeatureFlags = _VK_FORMAT_FEATURE_COLOR_ATTACHMENT_BLEND_BIT
	FORMAT_FEATURE_SAMPLED_IMAGE_BIT               FormatFeatureFlags = _VK_FORMAT_FEATURE_SAMPLED_IMAGE_BIT
	FORMAT_FEATURE_SAMPLED_IMAGE_FILTER_LINEAR_BIT FormatFeatureFlags = _VK_FORMAT_FEATURE_SAMPLED_IMAGE_FILTER_LINEAR_BIT

	IMAGE_USAGE_SAMPLED_BIT          ImageUsageFlags = _VK_IMAGE_USAGE_SAMPLED_BIT
	IMAGE_USAGE_COLOR_ATTACHMENT_BIT ImageUsageFlags = _VK_IMAGE_USAGE_COLOR_ATTACHMENT_BIT
	IMAGE_USAGE_STORAGE_BIT          ImageUsageFlags = _VK_IMAGE_USAGE_STORAGE_BIT
	IMAGE_USAGE_TRANSFER_DST_BIT     ImageUsageFlags = _VK_IMAGE_USAGE_TRANSFER_DST_BIT
	IMAGE_USAGE_TRANSFER_SRC_BIT     ImageUsageFlags = _VK_IMAGE_USAGE_TRANSFER_SRC_BIT

	FILTER_NEAREST Filter = _VK_FILTER_NEAREST
	FILTER_LINEAR  Filter = _VK_FILTER_LINEAR

	ATTACHMENT_LOAD_OP_CLEAR     AttachmentLoadOp = _VK_ATTACHMENT_LOAD_OP_CLEAR
	ATTACHMENT_LOAD_OP_DONT_CARE AttachmentLoadOp = _VK_ATTACHMENT_LOAD_OP_DONT_CARE
	ATTACHMENT_LOAD_OP_LOAD      AttachmentLoadOp = _VK_ATTACHMENT_LOAD_OP_LOAD

	IMAGE_LAYOUT_UNDEFINED                ImageLayout = _VK_IMAGE_LAYOUT_UNDEFINED
	IMAGE_LAYOUT_PRESENT_SRC_KHR          ImageLayout = _VK_IMAGE_LAYOUT_PRESENT_SRC_KHR
	IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL ImageLayout = _VK_IMAGE_LAYOUT_SHADER_READ_ONLY_OPTIMAL
	IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL ImageLayout = _VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL
	IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL     ImageLayout = _VK_IMAGE_LAYOUT_TRANSFER_DST_OPTIMAL
	IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL     ImageLayout = _VK_IMAGE_LAYOUT_TRANSFER_SRC_OPTIMAL
	IMAGE_LAYOUT_GENERAL                  ImageLayout = _VK_IMAGE_LAYOUT_GENERAL

	BUFFER_USAGE_TRANSFER_DST_BIT   BufferUsageFlags = _VK_BUFFER_USAGE_TRANSFER_DST_BIT
	BUFFER_USAGE_TRANSFER_SRC_BIT   BufferUsageFlags = _VK_BUFFER_USAGE_TRANSFER_SRC_BIT
	BUFFER_USAGE_UNIFORM_BUFFER_BIT BufferUsageFlags = _VK_BUFFER_USAGE_UNIFORM_BUFFER_BIT
	BUFFER_USAGE_STORAGE_BUFFER_BIT BufferUsageFlags = _VK_BUFFER_USAGE_STORAGE_BUFFER_BIT
	BUFFER_USAGE_INDEX_BUFFER_BIT   BufferUsageFlags = _VK_BUFFER_USAGE_INDEX_BUFFER_BIT
	BUFFER_USAGE_VERTEX_BUFFER_BIT  BufferUsageFlags = _VK_BUFFER_USAGE_VERTEX_BUFFER_BIT

	ERROR_OUT_OF_DATE_KHR  = Error(_VK_ERROR_OUT_OF_DATE_KHR)
	ERROR_SURFACE_LOST_KHR = Error(_VK_ERROR_SURFACE_LOST_KHR)
	ERROR_DEVICE_LOST      = Error(_VK_ERROR_DEVICE_LOST)
	SUBOPTIMAL_KHR         = Error(_VK_SUBOPTIMAL_KHR)

	FENCE_CREATE_SIGNALED_BIT = 0x00000001

	BLEND_FACTOR_ZERO                BlendFactor = _VK_BLEND_FACTOR_ZERO
	BLEND_FACTOR_ONE                 BlendFactor = _VK_BLEND_FACTOR_ONE
	BLEND_FACTOR_ONE_MINUS_SRC_ALPHA BlendFactor = _VK_BLEND_FACTOR_ONE_MINUS_SRC_ALPHA
	BLEND_FACTOR_DST_COLOR           BlendFactor = _VK_BLEND_FACTOR_DST_COLOR

	PRIMITIVE_TOPOLOGY_TRIANGLE_LIST  PrimitiveTopology = _VK_PRIMITIVE_TOPOLOGY_TRIANGLE_LIST
	PRIMITIVE_TOPOLOGY_TRIANGLE_STRIP PrimitiveTopology = _VK_PRIMITIVE_TOPOLOGY_TRIANGLE_STRIP

	SHADER_STAGE_VERTEX_BIT   ShaderStageFlags = _VK_SHADER_STAGE_VERTEX_BIT
	SHADER_STAGE_FRAGMENT_BIT ShaderStageFlags = _VK_SHADER_STAGE_FRAGMENT_BIT
	SHADER_STAGE_COMPUTE_BIT  ShaderStageFlags = _VK_SHADER_STAGE_COMPUTE_BIT

	DESCRIPTOR_TYPE_COMBINED_IMAGE_SAMPLER DescriptorType = _VK_DESCRIPTOR_TYPE_COMBINED_IMAGE_SAMPLER
	DESCRIPTOR_TYPE_UNIFORM_BUFFER         DescriptorType = _VK_DESCRIPTOR_TYPE_UNIFORM_BUFFER
	DESCRIPTOR_TYPE_STORAGE_BUFFER         DescriptorType = _VK_DESCRIPTOR_TYPE_STORAGE_BUFFER
	DESCRIPTOR_TYPE_STORAGE_IMAGE          DescriptorType = _VK_DESCRIPTOR_TYPE_STORAGE_IMAGE

	MEMORY_PROPERTY_DEVICE_LOCAL_BIT  MemoryPropertyFlags = _VK_MEMORY_PROPERTY_DEVICE_LOCAL_BIT
	MEMORY_PROPERTY_HOST_VISIBLE_BIT  MemoryPropertyFlags = _VK_MEMORY_PROPERTY_HOST_VISIBLE_BIT
	MEMORY_PROPERTY_HOST_COHERENT_BIT MemoryPropertyFlags = _VK_MEMORY_PROPERTY_HOST_COHERENT_BIT

	DEPENDENCY_BY_REGION_BIT DependencyFlags = _VK_DEPENDENCY_BY_REGION_BIT

	PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT PipelineStageFlags = _VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT
	PIPELINE_STAGE_TRANSFER_BIT                PipelineStageFlags = _VK_PIPELINE_STAGE_TRANSFER_BIT
	PIPELINE_STAGE_FRAGMENT_SHADER_BIT         PipelineStageFlags = _VK_PIPELINE_STAGE_FRAGMENT_SHADER_BIT
	PIPELINE_STAGE_COMPUTE_SHADER_BIT          PipelineStageFlags = _VK_PIPELINE_STAGE_COMPUTE_SHADER_BIT
	PIPELINE_STAGE_TOP_OF_PIPE_BIT             PipelineStageFlags = _VK_PIPELINE_STAGE_TOP_OF_PIPE_BIT
	PIPELINE_STAGE_HOST_BIT                    PipelineStageFlags = _VK_PIPELINE_STAGE_HOST_BIT
	PIPELINE_STAGE_VERTEX_INPUT_BIT            PipelineStageFlags = _VK_PIPELINE_STAGE_VERTEX_INPUT_BIT
	PIPELINE_STAGE_BOTTOM_OF_PIPE_BIT          PipelineStageFlags = _VK_PIPELINE_STAGE_BOTTOM_OF_PIPE_BIT

	ACCESS_MEMORY_READ_BIT            AccessFlags = _VK_ACCESS_MEMORY_READ_BIT
	ACCESS_MEMORY_WRITE_BIT           AccessFlags = _VK_ACCESS_MEMORY_WRITE_BIT
	ACCESS_TRANSFER_READ_BIT          AccessFlags = _VK_ACCESS_TRANSFER_READ_BIT
	ACCESS_TRANSFER_WRITE_BIT         AccessFlags = _VK_ACCESS_TRANSFER_WRITE_BIT
	ACCESS_SHADER_READ_BIT            AccessFlags = _VK_ACCESS_SHADER_READ_BIT
	ACCESS_SHADER_WRITE_BIT           AccessFlags = _VK_ACCESS_SHADER_WRITE_BIT
	ACCESS_COLOR_ATTACHMENT_READ_BIT  AccessFlags = _VK_ACCESS_COLOR_ATTACHMENT_READ_BIT
	ACCESS_COLOR_ATTACHMENT_WRITE_BIT AccessFlags = _VK_ACCESS_COLOR_ATTACHMENT_WRITE_BIT
	ACCESS_HOST_READ_BIT              AccessFlags = _VK_ACCESS_HOST_READ_BIT
	ACCESS_HOST_WRITE_BIT             AccessFlags = _VK_ACCESS_HOST_WRITE_BIT
	ACCESS_VERTEX_ATTRIBUTE_READ_BIT  AccessFlags = _VK_ACCESS_VERTEX_ATTRIBUTE_READ_BIT
	ACCESS_INDEX_READ_BIT             AccessFlags = _VK_ACCESS_INDEX_READ_BIT

	PIPELINE_BIND_POINT_COMPUTE  PipelineBindPoint = _VK_PIPELINE_BIND_POINT_COMPUTE
	PIPELINE_BIND_POINT_GRAPHICS PipelineBindPoint = _VK_PIPELINE_BIND_POINT_GRAPHICS

	INDEX_TYPE_UINT16 IndexType = _VK_INDEX_TYPE_UINT16
	INDEX_TYPE_UINT32 IndexType = _VK_INDEX_TYPE_UINT32

	QUEUE_GRAPHICS_BIT QueueFlags = _VK_QUEUE_GRAPHICS_BIT
	QUEUE_COMPUTE_BIT  QueueFlags = _VK_QUEUE_COMPUTE_BIT

	SAMPLER_MIPMAP_MODE_NEAREST SamplerMipmapMode = _VK_SAMPLER_MIPMAP_MODE_NEAREST
	SAMPLER_MIPMAP_MODE_LINEAR  SamplerMipmapMode = _VK_SAMPLER_MIPMAP_MODE_LINEAR

	REMAINING_MIP_LEVELS = -1
)

var (
	once    sync.Once
	loadErr error

	loadFuncs []func(dlopen func(name string) uintptr)
)

var funcs struct {
	vkCreateInstance                         uintptr
	vkDestroyInstance                        uintptr
	vkEnumeratePhysicalDevices               uintptr
	vkGetPhysicalDeviceQueueFamilyProperties uintptr
	vkGetPhysicalDeviceFormatProperties      uintptr
	vkCreateDevice                           uintptr
	vkDestroyDevice                          uintptr
	vkGetDeviceQueue                         uintptr
	vkCreateImageView                        uintptr
	vkDestroyImageView                       uintptr
	vkCreateFramebuffer                      uintptr
	vkDestroyFramebuffer                     uintptr
	vkDeviceWaitIdle                         uintptr
	vkQueueWaitIdle                          uintptr
	vkCreateSemaphore                        uintptr
	vkDestroySemaphore                       uintptr
	vkCreateRenderPass                       uintptr
	vkDestroyRenderPass                      uintptr
	vkCreateCommandPool                      uintptr
	vkDestroyCommandPool                     uintptr
	vkAllocateCommandBuffers                 uintptr
	vkFreeCommandBuffers                     uintptr
	vkBeginCommandBuffer                     uintptr
	vkEndCommandBuffer                       uintptr
	vkQueueSubmit                            uintptr
	vkCmdBeginRenderPass                     uintptr
	vkCmdEndRenderPass                       uintptr
	vkCmdCopyBuffer                          uintptr
	vkCmdCopyBufferToImage                   uintptr
	vkCmdPipelineBarrier                     uintptr
	vkCmdPushConstants                       uintptr
	vkCmdBindPipeline                        uintptr
	vkCmdBindVertexBuffers                   uintptr
	vkCmdSetViewport                         uintptr
	vkCmdBindIndexBuffer                     uintptr
	vkCmdDraw                                uintptr
	vkCmdDrawIndexed                         uintptr
	vkCmdBindDescriptorSets                  uintptr
	vkCmdCopyImageToBuffer                   uintptr
	vkCmdDispatch                            uintptr
	vkCreateImage                            uintptr
	vkDestroyImage                           uintptr
	vkGetImageMemoryRequirements             uintptr
	vkAllocateMemory                         uintptr
	vkBindImageMemory                        uintptr
	vkFreeMemory                             uintptr
	vkGetPhysicalDeviceMemoryProperties      uintptr
	vkCreateSampler                          uintptr
	vkDestroySampler                         uintptr
	vkCreateBuffer                           uintptr
	vkDestroyBuffer                          uintptr
	vkGetBufferMemoryRequirements            uintptr
	vkBindBufferMemory                       uintptr
	vkCreateShaderModule                     uintptr
	vkDestroyShaderModule                    uintptr
	vkCreateGraphicsPipelines                uintptr
	vkDestroyPipeline                        uintptr
	vkCreatePipelineLayout                   uintptr
	vkDestroyPipelineLayout                  uintptr
	vkCreateDescriptorSetLayout              uintptr
	vkDestroyDescriptorSetLayout             uintptr
	vkMapMemory                              uintptr
	vkUnmapMemory                            uintptr
	vkResetCommandBuffer                     uintptr
	vkCreateDescriptorPool                   uintptr
	vkDestroyDescriptorPool                  uintptr
	vkAllocateDescriptorSets                 uintptr
	vkFreeDescriptorSets                     uintptr
	vkUpdateDescriptorSets                   uintptr
	vkResetDescriptorPool                    uintptr
	vkCmdBlitImage                           uintptr
	vkCmdCopyImage                           uintptr
	vkCreateComputePipelines                 uintptr
	vkCreateFence                            uintptr
	vkDestroyFence                           uintptr
	vkWaitForFences                          uintptr
	vkResetFences                            uintptr
	vkGetPhysicalDeviceProperties            uintptr

	vkGetPhysicalDeviceSurfaceSupportKHR      uintptr
	vkDestroySurfaceKHR                       uintptr
	vkGetPhysicalDeviceSurfaceFormatsKHR      uintptr
	vkGetPhysicalDeviceSurfacePresentModesKHR uintptr
	vkGetPhysicalDeviceSurfaceCapabilitiesKHR uintptr

	vkCreateSwapchainKHR    uintptr
	vkDestroySwapchainKHR   uintptr
	vkGetSwapchainImagesKHR uintptr
	vkAcquireNextImageKHR   uintptr
	vkQueuePresentKHR       uintptr
}

var (
	nilSurface             vkSurfaceKHR
	nilSwapchain           vkSwapchainKHR
	nilSemaphore           vkSemaphore
	nilImageView           vkImageView
	nilRenderPass          vkRenderPass
	nilFramebuffer         vkFramebuffer
	nilCommandPool         vkCommandPool
	nilImage               vkImage
	nilDeviceMemory        vkDeviceMemory
	nilSampler             vkSampler
	nilBuffer              vkBuffer
	nilShaderModule        vkShaderModule
	nilPipeline            vkPipeline
	nilPipelineCache       vkPipelineCache
	nilPipelineLayout      vkPipelineLayout
	nilDescriptorSetLayout vkDescriptorSetLayout
	nilDescriptorPool      vkDescriptorPool
	nilDescriptorSet       vkDescriptorSet
	nilFence               vkFence
)

func vkInit() error {
	once.Do(func() {
		var libName string
		switch {
		case runtime.GOOS == "android":
			libName = "libvulkan.so"
		default:
			libName = "libvulkan.so.1"
		}
		lib, err := purego.Dlopen(libName, purego.RTLD_NOW|purego.RTLD_LOCAL)
		if err != nil {
			loadErr = fmt.Errorf("vulkan: %w", err)
			return
		}
		dlopen := func(name string) uintptr {
			ptr, _ := purego.Dlsym(lib, name)
			return ptr
		}
		must := func(name string) uintptr {
			ptr, err := purego.Dlsym(lib, name)
			if err == nil {
				return ptr
			}
			if loadErr == nil {
				loadErr = fmt.Errorf("vulkan: function %q not found: %w", name, err)
			}
			return 0
		}
		funcs.vkCreateInstance = must("vkCreateInstance")
		funcs.vkDestroyInstance = must("vkDestroyInstance")
		funcs.vkEnumeratePhysicalDevices = must("vkEnumeratePhysicalDevices")
		funcs.vkGetPhysicalDeviceQueueFamilyProperties = must("vkGetPhysicalDeviceQueueFamilyProperties")
		funcs.vkGetPhysicalDeviceFormatProperties = must("vkGetPhysicalDeviceFormatProperties")
		funcs.vkCreateDevice = must("vkCreateDevice")
		funcs.vkDestroyDevice = must("vkDestroyDevice")
		funcs.vkGetDeviceQueue = must("vkGetDeviceQueue")
		funcs.vkCreateImageView = must("vkCreateImageView")
		funcs.vkDestroyImageView = must("vkDestroyImageView")
		funcs.vkCreateFramebuffer = must("vkCreateFramebuffer")
		funcs.vkDestroyFramebuffer = must("vkDestroyFramebuffer")
		funcs.vkDeviceWaitIdle = must("vkDeviceWaitIdle")
		funcs.vkQueueWaitIdle = must("vkQueueWaitIdle")
		funcs.vkCreateSemaphore = must("vkCreateSemaphore")
		funcs.vkDestroySemaphore = must("vkDestroySemaphore")
		funcs.vkCreateRenderPass = must("vkCreateRenderPass")
		funcs.vkDestroyRenderPass = must("vkDestroyRenderPass")
		funcs.vkCreateCommandPool = must("vkCreateCommandPool")
		funcs.vkDestroyCommandPool = must("vkDestroyCommandPool")
		funcs.vkAllocateCommandBuffers = must("vkAllocateCommandBuffers")
		funcs.vkFreeCommandBuffers = must("vkFreeCommandBuffers")
		funcs.vkBeginCommandBuffer = must("vkBeginCommandBuffer")
		funcs.vkEndCommandBuffer = must("vkEndCommandBuffer")
		funcs.vkQueueSubmit = must("vkQueueSubmit")
		funcs.vkCmdBeginRenderPass = must("vkCmdBeginRenderPass")
		funcs.vkCmdEndRenderPass = must("vkCmdEndRenderPass")
		funcs.vkCmdCopyBuffer = must("vkCmdCopyBuffer")
		funcs.vkCmdCopyBufferToImage = must("vkCmdCopyBufferToImage")
		funcs.vkCmdPipelineBarrier = must("vkCmdPipelineBarrier")
		funcs.vkCmdPushConstants = must("vkCmdPushConstants")
		funcs.vkCmdBindPipeline = must("vkCmdBindPipeline")
		funcs.vkCmdBindVertexBuffers = must("vkCmdBindVertexBuffers")
		funcs.vkCmdSetViewport = must("vkCmdSetViewport")
		funcs.vkCmdBindIndexBuffer = must("vkCmdBindIndexBuffer")
		funcs.vkCmdDraw = must("vkCmdDraw")
		funcs.vkCmdDrawIndexed = must("vkCmdDrawIndexed")
		funcs.vkCmdBindDescriptorSets = must("vkCmdBindDescriptorSets")
		funcs.vkCmdCopyImageToBuffer = must("vkCmdCopyImageToBuffer")
		funcs.vkCmdDispatch = must("vkCmdDispatch")
		funcs.vkCreateImage = must("vkCreateImage")
		funcs.vkDestroyImage = must("vkDestroyImage")
		funcs.vkGetImageMemoryRequirements = must("vkGetImageMemoryRequirements")
		funcs.vkAllocateMemory = must("vkAllocateMemory")
		funcs.vkBindImageMemory = must("vkBindImageMemory")
		funcs.vkFreeMemory = must("vkFreeMemory")
		funcs.vkGetPhysicalDeviceMemoryProperties = must("vkGetPhysicalDeviceMemoryProperties")
		funcs.vkCreateSampler = must("vkCreateSampler")
		funcs.vkDestroySampler = must("vkDestroySampler")
		funcs.vkCreateBuffer = must("vkCreateBuffer")
		funcs.vkDestroyBuffer = must("vkDestroyBuffer")
		funcs.vkGetBufferMemoryRequirements = must("vkGetBufferMemoryRequirements")
		funcs.vkBindBufferMemory = must("vkBindBufferMemory")
		funcs.vkCreateShaderModule = must("vkCreateShaderModule")
		funcs.vkDestroyShaderModule = must("vkDestroyShaderModule")
		funcs.vkCreateGraphicsPipelines = must("vkCreateGraphicsPipelines")
		funcs.vkDestroyPipeline = must("vkDestroyPipeline")
		funcs.vkCreatePipelineLayout = must("vkCreatePipelineLayout")
		funcs.vkDestroyPipelineLayout = must("vkDestroyPipelineLayout")
		funcs.vkCreateDescriptorSetLayout = must("vkCreateDescriptorSetLayout")
		funcs.vkDestroyDescriptorSetLayout = must("vkDestroyDescriptorSetLayout")
		funcs.vkMapMemory = must("vkMapMemory")
		funcs.vkUnmapMemory = must("vkUnmapMemory")
		funcs.vkResetCommandBuffer = must("vkResetCommandBuffer")
		funcs.vkCreateDescriptorPool = must("vkCreateDescriptorPool")
		funcs.vkDestroyDescriptorPool = must("vkDestroyDescriptorPool")
		funcs.vkAllocateDescriptorSets = must("vkAllocateDescriptorSets")
		funcs.vkFreeDescriptorSets = must("vkFreeDescriptorSets")
		funcs.vkUpdateDescriptorSets = must("vkUpdateDescriptorSets")
		funcs.vkResetDescriptorPool = must("vkResetDescriptorPool")
		funcs.vkCmdBlitImage = must("vkCmdBlitImage")
		funcs.vkCmdCopyImage = must("vkCmdCopyImage")
		funcs.vkCreateComputePipelines = must("vkCreateComputePipelines")
		funcs.vkCreateFence = must("vkCreateFence")
		funcs.vkDestroyFence = must("vkDestroyFence")
		funcs.vkWaitForFences = must("vkWaitForFences")
		funcs.vkResetFences = must("vkResetFences")
		funcs.vkGetPhysicalDeviceProperties = must("vkGetPhysicalDeviceProperties")

		funcs.vkGetPhysicalDeviceSurfaceSupportKHR = dlopen("vkGetPhysicalDeviceSurfaceSupportKHR")
		funcs.vkDestroySurfaceKHR = dlopen("vkDestroySurfaceKHR")
		funcs.vkGetPhysicalDeviceSurfaceFormatsKHR = dlopen("vkGetPhysicalDeviceSurfaceFormatsKHR")
		funcs.vkGetPhysicalDeviceSurfacePresentModesKHR = dlopen("vkGetPhysicalDeviceSurfacePresentModesKHR")
		funcs.vkGetPhysicalDeviceSurfaceCapabilitiesKHR = dlopen("vkGetPhysicalDeviceSurfaceCapabilitiesKHR")

		funcs.vkCreateSwapchainKHR = dlopen("vkCreateSwapchainKHR")
		funcs.vkDestroySwapchainKHR = dlopen("vkDestroySwapchainKHR")
		funcs.vkGetSwapchainImagesKHR = dlopen("vkGetSwapchainImagesKHR")
		funcs.vkAcquireNextImageKHR = dlopen("vkAcquireNextImageKHR")
		funcs.vkQueuePresentKHR = dlopen("vkQueuePresentKHR")

		for _, f := range loadFuncs {
			f(dlopen)
		}
	})
	return loadErr
}

func CreateInstance(exts ...string) (Instance, error) {
	if err := vkInit(); err != nil {
		return nil, err
	}
	// VK_MAKE_API_VERSION macro.
	makeVer := func(variant, major, minor, patch int) uint32 {
		return ((((uint32)(variant)) << 29) | (((uint32)(major)) << 22) | (((uint32)(minor)) << 12) | ((uint32)(patch)))
	}
	appInf := vkApplicationInfo{
		sType:      _VK_STRUCTURE_TYPE_APPLICATION_INFO,
		apiVersion: makeVer(0, 1, 0, 0),
	}
	inf := vkInstanceCreateInfo{
		sType: _VK_STRUCTURE_TYPE_INSTANCE_CREATE_INFO,
		// pApplicationInfo may be omitted according to the spec, but the Android
		// emulator crashes without it.
		pApplicationInfo: &appInf,
	}
	if len(exts) > 0 {
		cexts := mallocCStringArr(exts)
		defer freeCStringArr(cexts)
		inf.enabledExtensionCount = uint32(len(exts))
		inf.ppEnabledExtensionNames = &cexts[0]
	}
	var inst Instance
	if err := vkErr(vkCreateInstance(funcs.vkCreateInstance, inf, nil, &inst)); err != nil {
		return nil, fmt.Errorf("vulkan: vkCreateInstance: %w", err)
	}
	return inst, nil
}

func mallocCStringArr(s []string) []*byte {
	carr := make([]*byte, len(s))
	for i, ext := range s {
		carr[i] = cString(ext)
	}
	return carr
}

func freeCStringArr(s []*byte) {
	for i := range s {
		cFree(unsafe.Pointer(s[i]))
		s[i] = nil
	}
}

func DestroyInstance(inst Instance) {
	vkDestroyInstance(funcs.vkDestroyInstance, inst, nil)
}

func GetPhysicalDeviceQueueFamilyProperties(pd PhysicalDevice) []QueueFamilyProperties {
	var count uint32
	vkGetPhysicalDeviceQueueFamilyProperties(funcs.vkGetPhysicalDeviceQueueFamilyProperties, pd, &count, nil)
	if count == 0 {
		return nil
	}
	queues := make([]vkQueueFamilyProperties, count)
	vkGetPhysicalDeviceQueueFamilyProperties(funcs.vkGetPhysicalDeviceQueueFamilyProperties, pd, &count, &queues[0])
	return queues
}

func EnumeratePhysicalDevices(inst Instance) ([]PhysicalDevice, error) {
	var count uint32
	if err := vkErr(vkEnumeratePhysicalDevices(funcs.vkEnumeratePhysicalDevices, inst, &count, nil)); err != nil {
		return nil, fmt.Errorf("vulkan: vkEnumeratePhysicalDevices: %w", err)
	}
	if count == 0 {
		return nil, nil
	}
	devs := make([]vkPhysicalDevice, count)
	if err := vkErr(vkEnumeratePhysicalDevices(funcs.vkEnumeratePhysicalDevices, inst, &count, &devs[0])); err != nil {
		return nil, fmt.Errorf("vulkan: vkEnumeratePhysicalDevices: %w", err)
	}
	return devs, nil
}

func ChoosePhysicalDevice(inst Instance, surf Surface) (PhysicalDevice, int, error) {
	devs, err := EnumeratePhysicalDevices(inst)
	if err != nil {
		return nil, 0, err
	}
	for _, pd := range devs {
		var props vkPhysicalDeviceProperties
		vkGetPhysicalDeviceProperties(funcs.vkGetPhysicalDeviceProperties, pd, &props)
		// The lavapipe software implementation doesn't work well rendering to a surface.
		// See https://gitlab.freedesktop.org/mesa/mesa/-/issues/5473.
		if surf != 0 && props.deviceType == _VK_PHYSICAL_DEVICE_TYPE_CPU {
			continue
		}
		const caps = _VK_QUEUE_GRAPHICS_BIT | _VK_QUEUE_COMPUTE_BIT
		queueIdx, ok, err := chooseQueue(pd, surf, caps)
		if err != nil {
			return nil, 0, err
		}
		if !ok {
			continue
		}
		if surf != nilSurface {
			_, fmtFound, err := chooseFormat(pd, surf)
			if err != nil {
				return nil, 0, err
			}
			_, modFound, err := choosePresentMode(pd, surf)
			if err != nil {
				return nil, 0, err
			}
			if !fmtFound || !modFound {
				continue
			}
		}
		return pd, queueIdx, nil
	}
	return nil, 0, errors.New("vulkan: no suitable device found")
}

func CreateDeviceAndQueue(pd vkPhysicalDevice, queueIdx int, exts ...string) (Device, error) {
	priority := float32(1.0)
	qinf := vkDeviceQueueCreateInfo{
		sType:            _VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO,
		queueCount:       1,
		queueFamilyIndex: uint32(queueIdx),
		pQueuePriorities: &priority,
	}
	inf := vkDeviceCreateInfo{
		sType:                 _VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO,
		queueCreateInfoCount:  1,
		enabledExtensionCount: uint32(len(exts)),
	}
	if len(exts) > 0 {
		cexts := mallocCStringArr(exts)
		defer freeCStringArr(cexts)
		inf.ppEnabledExtensionNames = &cexts[0]
	}
	var dev Device
	if err := vkErr(vkCreateDevice(funcs.vkCreateDevice, pd, inf, qinf, nil, &dev)); err != nil {
		return nil, fmt.Errorf("vulkan: vkCreateDevice: %w", err)
	}
	return dev, nil
}

func GetDeviceQueue(d Device, queueFamily, queueIndex int) Queue {
	var queue Queue
	vkGetDeviceQueue(funcs.vkGetDeviceQueue, d, uint32(queueFamily), uint32(queueIndex), &queue)
	return queue
}

func GetPhysicalDeviceSurfaceCapabilities(pd PhysicalDevice, surf Surface) (SurfaceCapabilities, error) {
	var caps vkSurfaceCapabilitiesKHR
	err := vkErr(vkGetPhysicalDeviceSurfaceCapabilitiesKHR(funcs.vkGetPhysicalDeviceSurfaceCapabilitiesKHR, pd, surf, &caps))
	if err != nil {
		return SurfaceCapabilities{}, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceCapabilitiesKHR: %w", err)
	}
	return caps, nil
}

func CreateSwapchain(pd PhysicalDevice, d Device, surf Surface, width, height int, old Swapchain) (Swapchain, []Image, Format, error) {
	caps, err := GetPhysicalDeviceSurfaceCapabilities(pd, surf)
	if err != nil {
		return nilSwapchain, nil, 0, err
	}
	mode, modeOK, err := choosePresentMode(pd, surf)
	if err != nil {
		return nilSwapchain, nil, 0, err
	}
	format, fmtOK, err := chooseFormat(pd, surf)
	if err != nil {
		return nilSwapchain, nil, 0, err
	}
	if !modeOK || !fmtOK {
		// This shouldn't happen because CreateDeviceAndQueue found at least
		// one valid format and present mode.
		return nilSwapchain, nil, 0, errors.New("vulkan: no valid format and present mode found")
	}
	// Find supported alpha composite mode. It doesn't matter which one, because rendering is
	// always opaque.
	alphaComp := vkCompositeAlphaFlagBitsKHR(1)
	for caps.supportedCompositeAlpha&vkCompositeAlphaFlagsKHR(alphaComp) == 0 {
		alphaComp <<= 1
	}
	trans := vkSurfaceTransformFlagBitsKHR(_VK_SURFACE_TRANSFORM_IDENTITY_BIT_KHR)
	if caps.supportedTransforms&vkSurfaceTransformFlagsKHR(trans) == 0 {
		return nilSwapchain, nil, 0, errors.New("vulkan: VK_SURFACE_TRANSFORM_IDENTITY_BIT_KHR not supported")
	}
	inf := vkSwapchainCreateInfoKHR{
		sType:            _VK_STRUCTURE_TYPE_SWAPCHAIN_CREATE_INFO_KHR,
		surface:          surf,
		minImageCount:    caps.minImageCount,
		imageFormat:      format.format,
		imageColorSpace:  format.colorSpace,
		imageExtent:      vkExtent2D{width: uint32(width), height: uint32(height)},
		imageArrayLayers: 1,
		imageUsage:       _VK_IMAGE_USAGE_COLOR_ATTACHMENT_BIT,
		imageSharingMode: _VK_SHARING_MODE_EXCLUSIVE,
		preTransform:     trans,
		presentMode:      mode,
		compositeAlpha:   vkCompositeAlphaFlagBitsKHR(alphaComp),
		clipped:          _VK_TRUE,
		oldSwapchain:     old,
	}
	var swchain Swapchain
	if err := vkErr(vkCreateSwapchainKHR(funcs.vkCreateSwapchainKHR, d, &inf, nil, &swchain)); err != nil {
		return nilSwapchain, nil, 0, fmt.Errorf("vulkan: vkCreateSwapchainKHR: %w", err)
	}
	var count uint32
	if err := vkErr(vkGetSwapchainImagesKHR(funcs.vkGetSwapchainImagesKHR, d, swchain, &count, nil)); err != nil {
		DestroySwapchain(d, swchain)
		return nilSwapchain, nil, 0, fmt.Errorf("vulkan: vkGetSwapchainImagesKHR: %w", err)
	}
	if count == 0 {
		DestroySwapchain(d, swchain)
		return nilSwapchain, nil, 0, errors.New("vulkan: vkGetSwapchainImagesKHR returned no images")
	}
	imgs := make([]Image, count)
	if err := vkErr(vkGetSwapchainImagesKHR(funcs.vkGetSwapchainImagesKHR, d, swchain, &count, &imgs[0])); err != nil {
		DestroySwapchain(d, swchain)
		return nilSwapchain, nil, 0, fmt.Errorf("vulkan: vkGetSwapchainImagesKHR: %w", err)
	}
	return swchain, imgs, format.format, nil
}

func DestroySwapchain(d Device, swchain Swapchain) {
	vkDestroySwapchainKHR(funcs.vkDestroySwapchainKHR, d, swchain, nil)
}

func AcquireNextImage(d Device, swchain Swapchain, sem Semaphore, fence Fence) (int, error) {
	res := vkAcquireNextImageKHR(funcs.vkAcquireNextImageKHR, d, swchain, math.MaxUint64, sem, fence)
	if err := vkErr(res.res); err != nil {
		return 0, fmt.Errorf("vulkan: vkAcquireNextImageKHR: %w", err)
	}
	return int(res.uint), nil
}

func PresentQueue(q Queue, swchain Swapchain, sem Semaphore, imgIdx int) error {
	cidx := uint32(imgIdx)
	inf := vkPresentInfoKHR{
		sType:          _VK_STRUCTURE_TYPE_PRESENT_INFO_KHR,
		swapchainCount: 1,
		pSwapchains:    &swchain,
		pImageIndices:  &cidx,
	}
	if sem != nilSemaphore {
		inf.waitSemaphoreCount = 1
		inf.pWaitSemaphores = &sem
	}
	if err := vkErr(vkQueuePresentKHR(funcs.vkQueuePresentKHR, q, inf)); err != nil {
		return fmt.Errorf("vulkan: vkQueuePresentKHR: %w", err)
	}
	return nil
}

func CreateImageView(d Device, img Image, format Format) (ImageView, error) {
	inf := vkImageViewCreateInfo{
		sType:    _VK_STRUCTURE_TYPE_IMAGE_VIEW_CREATE_INFO,
		image:    img,
		viewType: _VK_IMAGE_VIEW_TYPE_2D,
		format:   format,
		subresourceRange: vkImageSubresourceRange{
			aspectMask: _VK_IMAGE_ASPECT_COLOR_BIT,
			levelCount: _VK_REMAINING_MIP_LEVELS,
			layerCount: _VK_REMAINING_ARRAY_LAYERS,
		},
	}
	var view vkImageView
	if err := vkErr(vkCreateImageView(funcs.vkCreateImageView, d, &inf, nil, &view)); err != nil {
		return nilImageView, fmt.Errorf("vulkan: vkCreateImageView: %w", err)
	}
	return view, nil
}

func DestroyImageView(d Device, view ImageView) {
	vkDestroyImageView(funcs.vkDestroyImageView, d, view, nil)
}

func CreateRenderPass(d Device, format Format, loadOp AttachmentLoadOp, initialLayout, finalLayout ImageLayout, passDeps []SubpassDependency) (RenderPass, error) {
	att := vkAttachmentDescription{
		format:         format,
		samples:        _VK_SAMPLE_COUNT_1_BIT,
		loadOp:         loadOp,
		storeOp:        _VK_ATTACHMENT_STORE_OP_STORE,
		stencilLoadOp:  _VK_ATTACHMENT_LOAD_OP_DONT_CARE,
		stencilStoreOp: _VK_ATTACHMENT_STORE_OP_DONT_CARE,
		initialLayout:  initialLayout,
		finalLayout:    finalLayout,
	}

	ref := vkAttachmentReference{
		attachment: 0,
		layout:     _VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL,
	}

	sub := vkSubpassDescription{
		pipelineBindPoint:    _VK_PIPELINE_BIND_POINT_GRAPHICS,
		colorAttachmentCount: 1,
		pColorAttachments:    &ref,
	}

	inf := vkRenderPassCreateInfo{
		sType:           _VK_STRUCTURE_TYPE_RENDER_PASS_CREATE_INFO,
		attachmentCount: 1,
		pAttachments:    &att,
		subpassCount:    1,
	}
	if n := len(passDeps); n > 0 {
		inf.dependencyCount = uint32(n)
		inf.pDependencies = &passDeps[0]
	}

	var pass RenderPass
	if err := vkErr(vkCreateRenderPass(funcs.vkCreateRenderPass, d, inf, sub, nil, &pass)); err != nil {
		return nilRenderPass, fmt.Errorf("vulkan: vkCreateRenderPass: %w", err)
	}
	return pass, nil
}

func DestroyRenderPass(d Device, r RenderPass) {
	vkDestroyRenderPass(funcs.vkDestroyRenderPass, d, r, nil)
}

func CreateFramebuffer(d Device, rp RenderPass, view ImageView, width, height int) (Framebuffer, error) {
	inf := vkFramebufferCreateInfo{
		sType:           _VK_STRUCTURE_TYPE_FRAMEBUFFER_CREATE_INFO,
		renderPass:      rp,
		attachmentCount: 1,
		pAttachments:    &view,
		width:           uint32(width),
		height:          uint32(height),
		layers:          1,
	}
	var fbo Framebuffer
	if err := vkErr(vkCreateFramebuffer(funcs.vkCreateFramebuffer, d, inf, nil, &fbo)); err != nil {
		return nilFramebuffer, fmt.Errorf("vulkan: vkCreateFramebuffer: %w", err)
	}
	return fbo, nil
}

func DestroyFramebuffer(d Device, f Framebuffer) {
	vkDestroyFramebuffer(funcs.vkDestroyFramebuffer, d, f, nil)
}

func DeviceWaitIdle(d Device) error {
	if err := vkErr(vkDeviceWaitIdle(funcs.vkDeviceWaitIdle, d)); err != nil {
		return fmt.Errorf("vulkan: vkDeviceWaitIdle: %w", err)
	}
	return nil
}

func QueueWaitIdle(q Queue) error {
	if err := vkErr(vkQueueWaitIdle(funcs.vkQueueWaitIdle, q)); err != nil {
		return fmt.Errorf("vulkan: vkQueueWaitIdle: %w", err)
	}
	return nil
}

func CreateSemaphore(d Device) (Semaphore, error) {
	inf := vkSemaphoreCreateInfo{
		sType: _VK_STRUCTURE_TYPE_SEMAPHORE_CREATE_INFO,
	}
	var sem Semaphore
	err := vkErr(vkCreateSemaphore(funcs.vkCreateSemaphore, d, &inf, nil, &sem))
	if err != nil {
		return nilSemaphore, fmt.Errorf("vulkan: vkCreateSemaphore: %w", err)
	}
	return sem, err
}

func DestroySemaphore(d Device, sem Semaphore) {
	vkDestroySemaphore(funcs.vkDestroySemaphore, d, sem, nil)
}

func DestroyDevice(dev Device) {
	vkDestroyDevice(funcs.vkDestroyDevice, dev, nil)
}

func DestroySurface(inst Instance, s Surface) {
	vkDestroySurfaceKHR(funcs.vkDestroySurfaceKHR, inst, s, nil)
}

func CreateCommandPool(d Device, queueIndex int) (CommandPool, error) {
	inf := vkCommandPoolCreateInfo{
		sType:            _VK_STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO,
		queueFamilyIndex: uint32(queueIndex),
		flags:            _VK_COMMAND_POOL_CREATE_TRANSIENT_BIT | _VK_COMMAND_POOL_CREATE_RESET_COMMAND_BUFFER_BIT,
	}

	var pool CommandPool
	if err := vkErr(vkCreateCommandPool(funcs.vkCreateCommandPool, d, &inf, nil, &pool)); err != nil {
		return nilCommandPool, fmt.Errorf("vulkan: vkCreateCommandPool: %w", err)
	}
	return pool, nil
}

func DestroyCommandPool(d Device, pool CommandPool) {
	vkDestroyCommandPool(funcs.vkDestroyCommandPool, d, pool, nil)
}

func AllocateCommandBuffer(d Device, pool CommandPool) (CommandBuffer, error) {
	inf := vkCommandBufferAllocateInfo{
		sType:              _VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO,
		commandPool:        pool,
		level:              _VK_COMMAND_BUFFER_LEVEL_PRIMARY,
		commandBufferCount: 1,
	}

	var buf CommandBuffer
	if err := vkErr(vkAllocateCommandBuffers(funcs.vkAllocateCommandBuffers, d, &inf, &buf)); err != nil {
		return nil, fmt.Errorf("vulkan: vkAllocateCommandBuffers: %w", err)
	}
	return buf, nil
}

func FreeCommandBuffers(d Device, pool CommandPool, bufs ...CommandBuffer) {
	if len(bufs) == 0 {
		return
	}
	vkFreeCommandBuffers(funcs.vkFreeCommandBuffers, d, pool, uint32(len(bufs)), &bufs[0])
}

func BeginCommandBuffer(buf CommandBuffer) error {
	inf := vkCommandBufferBeginInfo{
		sType: _VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO,
		flags: _VK_COMMAND_BUFFER_USAGE_ONE_TIME_SUBMIT_BIT,
	}
	if err := vkErr(vkBeginCommandBuffer(funcs.vkBeginCommandBuffer, buf, inf)); err != nil {
		return fmt.Errorf("vulkan: vkBeginCommandBuffer: %w", err)
	}
	return nil
}

func EndCommandBuffer(buf CommandBuffer) error {
	if err := vkErr(vkEndCommandBuffer(funcs.vkEndCommandBuffer, buf)); err != nil {
		return fmt.Errorf("vulkan: vkEndCommandBuffer: %w", err)
	}
	return nil
}

func QueueSubmit(q Queue, buf CommandBuffer, waitSems []Semaphore, waitStages []PipelineStageFlags, sigSems []Semaphore, fence Fence) error {
	inf := vkSubmitInfo{
		sType:              _VK_STRUCTURE_TYPE_SUBMIT_INFO,
		commandBufferCount: 1,
		pCommandBuffers:    &buf,
	}
	if len(waitSems) > 0 {
		if len(waitSems) != len(waitStages) {
			panic("len(waitSems) != len(waitStages)")
		}
		inf.waitSemaphoreCount = uint32(len(waitSems))
		inf.pWaitSemaphores = &waitSems[0]
		inf.pWaitDstStageMask = &waitStages[0]
	}
	if len(sigSems) > 0 {
		inf.signalSemaphoreCount = uint32(len(sigSems))
		inf.pSignalSemaphores = &sigSems[0]
	}
	if err := vkErr(vkQueueSubmit(funcs.vkQueueSubmit, q, inf, fence)); err != nil {
		return fmt.Errorf("vulkan: vkQueueSubmit: %w", err)
	}
	return nil
}

func CmdBeginRenderPass(buf CommandBuffer, rp RenderPass, fbo Framebuffer, width, height int, clearCol [4]float32) {
	cclearCol := [4]float32{float32(clearCol[0]), float32(clearCol[1]), float32(clearCol[2]), float32(clearCol[3])}
	inf := vkRenderPassBeginInfo{
		sType:           _VK_STRUCTURE_TYPE_RENDER_PASS_BEGIN_INFO,
		renderPass:      rp,
		framebuffer:     fbo,
		renderArea:      vkRect2D{extent: vkExtent2D{width: uint32(width), height: uint32(height)}},
		clearValueCount: 1,
		pClearValues:    (*vkClearValue)(unsafe.Pointer(&cclearCol)),
	}
	vkCmdBeginRenderPass(funcs.vkCmdBeginRenderPass, buf, inf, _VK_SUBPASS_CONTENTS_INLINE)
}

func CmdEndRenderPass(buf CommandBuffer) {
	vkCmdEndRenderPass(funcs.vkCmdEndRenderPass, buf)
}

func CmdCopyBuffer(cmdBuf CommandBuffer, src, dst Buffer, srcOff, dstOff, size int) {
	vkCmdCopyBuffer(funcs.vkCmdCopyBuffer, cmdBuf, src, dst, 1, &vkBufferCopy{
		srcOffset: vkDeviceSize(srcOff),
		dstOffset: vkDeviceSize(dstOff),
		size:      vkDeviceSize(size),
	})
}

func CmdCopyBufferToImage(cmdBuf CommandBuffer, src Buffer, dst Image, layout ImageLayout, copy BufferImageCopy) {
	vkCmdCopyBufferToImage(funcs.vkCmdCopyBufferToImage, cmdBuf, src, dst, layout, 1, &copy)
}

func CmdPipelineBarrier(cmdBuf CommandBuffer, srcStage, dstStage PipelineStageFlags, flags DependencyFlags, memBarriers []MemoryBarrier, bufBarriers []BufferMemoryBarrier, imgBarriers []ImageMemoryBarrier) {
	var memPtr *MemoryBarrier
	if len(memBarriers) > 0 {
		memPtr = &memBarriers[0]
	}
	var bufPtr *BufferMemoryBarrier
	if len(bufBarriers) > 0 {
		bufPtr = &bufBarriers[0]
	}
	var imgPtr *ImageMemoryBarrier
	if len(imgBarriers) > 0 {
		imgPtr = &imgBarriers[0]
	}
	vkCmdPipelineBarrier(funcs.vkCmdPipelineBarrier, cmdBuf, srcStage, dstStage, flags,
		uint32(len(memBarriers)), memPtr,
		uint32(len(bufBarriers)), bufPtr,
		uint32(len(imgBarriers)), imgPtr)
}

func CmdPushConstants(cmdBuf CommandBuffer, layout PipelineLayout, stages ShaderStageFlags, offset int, data []byte) {
	vkCmdPushConstants(funcs.vkCmdPushConstants, cmdBuf, layout, stages, uint32(offset), uint32(len(data)),
		unsafe.Pointer(unsafe.SliceData(data)))
}

func CmdBindPipeline(cmdBuf CommandBuffer, bindPoint PipelineBindPoint, pipe Pipeline) {
	vkCmdBindPipeline(funcs.vkCmdBindPipeline, cmdBuf, bindPoint, pipe)
}

func CmdBindVertexBuffers(cmdBuf CommandBuffer, first int, buffers []Buffer, sizes []DeviceSize) {
	if len(buffers) == 0 {
		return
	}
	vkCmdBindVertexBuffers(funcs.vkCmdBindVertexBuffers, cmdBuf, uint32(first), uint32(len(buffers)), &buffers[0], &sizes[0])
}

func CmdSetViewport(cmdBuf CommandBuffer, first int, viewports ...Viewport) {
	if len(viewports) == 0 {
		return
	}
	vkCmdSetViewport(funcs.vkCmdSetViewport, cmdBuf, uint32(first), uint32(len(viewports)), &viewports[0])
}

func CmdBindIndexBuffer(cmdBuf CommandBuffer, buffer Buffer, offset int, typ IndexType) {
	vkCmdBindIndexBuffer(funcs.vkCmdBindIndexBuffer, cmdBuf, buffer, vkDeviceSize(offset), typ)
}

func CmdDraw(cmdBuf CommandBuffer, vertCount, instCount, firstVert, firstInst int) {
	vkCmdDraw(funcs.vkCmdDraw, cmdBuf, uint32(vertCount), uint32(instCount), uint32(firstVert), uint32(firstInst))
}

func CmdDrawIndexed(cmdBuf CommandBuffer, idxCount, instCount, firstIdx, vertOff, firstInst int) {
	vkCmdDrawIndexed(funcs.vkCmdDrawIndexed, cmdBuf, uint32(idxCount), uint32(instCount), uint32(firstIdx), int32(vertOff), uint32(firstInst))
}

func GetPhysicalDeviceFormatProperties(physDev PhysicalDevice, format Format) FormatFeatureFlags {
	var props vkFormatProperties
	vkGetPhysicalDeviceFormatProperties(funcs.vkGetPhysicalDeviceFormatProperties, physDev, format, &props)
	return FormatFeatureFlags(props.optimalTilingFeatures)
}

func CmdBindDescriptorSets(cmdBuf CommandBuffer, point PipelineBindPoint, layout PipelineLayout, firstSet int, sets []DescriptorSet) {
	vkCmdBindDescriptorSets(funcs.vkCmdBindDescriptorSets, cmdBuf, point, layout, uint32(firstSet), uint32(len(sets)), &sets[0], 0, nil)
}

func CmdBlitImage(cmdBuf CommandBuffer, src Image, srcLayout ImageLayout, dst Image, dstLayout ImageLayout, regions []ImageBlit, filter Filter) {
	if len(regions) == 0 {
		return
	}
	vkCmdBlitImage(funcs.vkCmdBlitImage, cmdBuf, src, srcLayout, dst, dstLayout, uint32(len(regions)), &regions[0], filter)
}

func CmdCopyImage(cmdBuf CommandBuffer, src Image, srcLayout ImageLayout, dst Image, dstLayout ImageLayout, regions []ImageCopy) {
	if len(regions) == 0 {
		return
	}
	vkCmdCopyImage(funcs.vkCmdCopyImage, cmdBuf, src, srcLayout, dst, dstLayout, uint32(len(regions)), &regions[0])
}

func CmdCopyImageToBuffer(cmdBuf CommandBuffer, src Image, srcLayout ImageLayout, dst Buffer, regions []BufferImageCopy) {
	if len(regions) == 0 {
		return
	}
	vkCmdCopyImageToBuffer(funcs.vkCmdCopyImageToBuffer, cmdBuf, src, srcLayout, dst, uint32(len(regions)), &regions[0])
}

func CmdDispatch(cmdBuf CommandBuffer, x, y, z int) {
	vkCmdDispatch(funcs.vkCmdDispatch, cmdBuf, uint32(x), uint32(y), uint32(z))
}

func CreateImage(pd PhysicalDevice, d Device, format Format, width, height, mipmaps int, usage ImageUsageFlags) (Image, DeviceMemory, error) {
	inf := vkImageCreateInfo{
		sType:     _VK_STRUCTURE_TYPE_IMAGE_CREATE_INFO,
		imageType: _VK_IMAGE_TYPE_2D,
		format:    format,
		extent: vkExtent3D{
			width:  uint32(width),
			height: uint32(height),
			depth:  1,
		},
		mipLevels:     uint32(mipmaps),
		arrayLayers:   1,
		samples:       _VK_SAMPLE_COUNT_1_BIT,
		tiling:        _VK_IMAGE_TILING_OPTIMAL,
		usage:         usage,
		initialLayout: _VK_IMAGE_LAYOUT_UNDEFINED,
	}
	var img vkImage
	if err := vkErr(vkCreateImage(funcs.vkCreateImage, d, &inf, nil, &img)); err != nil {
		return nilImage, nilDeviceMemory, fmt.Errorf("vulkan: vkCreateImage: %w", err)
	}
	var memReqs vkMemoryRequirements
	vkGetImageMemoryRequirements(funcs.vkGetImageMemoryRequirements, d, img, &memReqs)

	memIdx, found := findMemoryTypeIndex(pd, memReqs.memoryTypeBits, _VK_MEMORY_PROPERTY_DEVICE_LOCAL_BIT)
	if !found {
		DestroyImage(d, img)
		return nilImage, nilDeviceMemory, errors.New("vulkan: no memory type suitable for images found")
	}

	memInf := vkMemoryAllocateInfo{
		sType:           _VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO,
		allocationSize:  memReqs.size,
		memoryTypeIndex: uint32(memIdx),
	}
	var imgMem vkDeviceMemory
	if err := vkErr(vkAllocateMemory(funcs.vkAllocateMemory, d, &memInf, nil, &imgMem)); err != nil {
		DestroyImage(d, img)
		return nilImage, nilDeviceMemory, fmt.Errorf("vulkan: vkAllocateMemory: %w", err)
	}

	if err := vkErr(vkBindImageMemory(funcs.vkBindImageMemory, d, img, imgMem, 0)); err != nil {
		FreeMemory(d, imgMem)
		DestroyImage(d, img)
		return nilImage, nilDeviceMemory, fmt.Errorf("vulkan: vkBindImageMemory: %w", err)
	}
	return img, imgMem, nil
}

func DestroyImage(d Device, img Image) {
	vkDestroyImage(funcs.vkDestroyImage, d, img, nil)
}

func FreeMemory(d Device, mem DeviceMemory) {
	vkFreeMemory(funcs.vkFreeMemory, d, mem, nil)
}

func CreateSampler(d Device, minFilter, magFilter Filter, mipmapMode SamplerMipmapMode) (Sampler, error) {
	inf := vkSamplerCreateInfo{
		sType:        _VK_STRUCTURE_TYPE_SAMPLER_CREATE_INFO,
		minFilter:    minFilter,
		magFilter:    magFilter,
		mipmapMode:   mipmapMode,
		maxLod:       _VK_LOD_CLAMP_NONE,
		addressModeU: _VK_SAMPLER_ADDRESS_MODE_CLAMP_TO_EDGE,
		addressModeV: _VK_SAMPLER_ADDRESS_MODE_CLAMP_TO_EDGE,
	}
	var s vkSampler
	if err := vkErr(vkCreateSampler(funcs.vkCreateSampler, d, &inf, nil, &s)); err != nil {
		return nilSampler, fmt.Errorf("vulkan: vkCreateSampler: %w", err)
	}
	return s, nil
}

func DestroySampler(d Device, sampler Sampler) {
	vkDestroySampler(funcs.vkDestroySampler, d, sampler, nil)
}

func CreateBuffer(pd PhysicalDevice, d Device, size int, usage BufferUsageFlags, props MemoryPropertyFlags) (Buffer, DeviceMemory, error) {
	inf := vkBufferCreateInfo{
		sType: _VK_STRUCTURE_TYPE_BUFFER_CREATE_INFO,
		size:  vkDeviceSize(size),
		usage: usage,
	}
	var buf vkBuffer
	if err := vkErr(vkCreateBuffer(funcs.vkCreateBuffer, d, &inf, nil, &buf)); err != nil {
		return nilBuffer, nilDeviceMemory, fmt.Errorf("vulkan: vkCreateBuffer: %w", err)
	}

	var memReqs vkMemoryRequirements
	vkGetBufferMemoryRequirements(funcs.vkGetBufferMemoryRequirements, d, buf, &memReqs)

	memIdx, found := findMemoryTypeIndex(pd, memReqs.memoryTypeBits, props)
	if !found {
		DestroyBuffer(d, buf)
		return nilBuffer, nilDeviceMemory, errors.New("vulkan: no memory suitable for buffers found")
	}
	memInf := vkMemoryAllocateInfo{
		sType:           _VK_STRUCTURE_TYPE_MEMORY_ALLOCATE_INFO,
		allocationSize:  memReqs.size,
		memoryTypeIndex: uint32(memIdx),
	}

	var mem vkDeviceMemory
	if err := vkErr(vkAllocateMemory(funcs.vkAllocateMemory, d, &memInf, nil, &mem)); err != nil {
		DestroyBuffer(d, buf)
		return nilBuffer, nilDeviceMemory, fmt.Errorf("vulkan: vkAllocateMemory: %w", err)
	}

	if err := vkErr(vkBindBufferMemory(funcs.vkBindBufferMemory, d, buf, mem, 0)); err != nil {
		FreeMemory(d, mem)
		DestroyBuffer(d, buf)
		return nilBuffer, nilDeviceMemory, fmt.Errorf("vulkan: vkBindBufferMemory: %w", err)
	}
	return buf, mem, nil
}

func DestroyBuffer(d Device, buf Buffer) {
	vkDestroyBuffer(funcs.vkDestroyBuffer, d, buf, nil)
}

func CreateShaderModule(d Device, spirv string) (ShaderModule, error) {
	inf := vkShaderModuleCreateInfo{
		sType:    _VK_STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO,
		codeSize: uint64(len(spirv)),
		pCode:    (*uint32)(unsafe.Pointer(unsafe.StringData(spirv))),
	}

	var mod vkShaderModule
	if err := vkErr(vkCreateShaderModule(funcs.vkCreateShaderModule, d, inf, nil, &mod)); err != nil {
		return nilShaderModule, fmt.Errorf("vulkan: vkCreateShaderModule: %w", err)
	}
	return mod, nil
}

func DestroyShaderModule(d Device, mod ShaderModule) {
	vkDestroyShaderModule(funcs.vkDestroyShaderModule, d, mod, nil)
}

func CreateGraphicsPipeline(d Device, pass RenderPass, vmod, fmod ShaderModule, blend bool, srcFactor, dstFactor BlendFactor, topology PrimitiveTopology, bindings []VertexInputBindingDescription, attrs []VertexInputAttributeDescription, layout PipelineLayout) (Pipeline, error) {
	main := cString("main")
	defer cFree(unsafe.Pointer(main))
	stages := []vkPipelineShaderStageCreateInfo{
		{
			sType:  _VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
			stage:  _VK_SHADER_STAGE_VERTEX_BIT,
			module: vmod,
			pName:  main,
		},
		{
			sType:  _VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
			stage:  _VK_SHADER_STAGE_FRAGMENT_BIT,
			module: fmod,
			pName:  main,
		},
	}
	dynStates := []vkDynamicState{_VK_DYNAMIC_STATE_VIEWPORT}
	dynInf := vkPipelineDynamicStateCreateInfo{
		sType:             _VK_STRUCTURE_TYPE_PIPELINE_DYNAMIC_STATE_CREATE_INFO,
		dynamicStateCount: uint32(len(dynStates)),
		pDynamicStates:    &dynStates[0],
	}
	const maxDim = 0x7fffffff
	scissors := []vkRect2D{{extent: vkExtent2D{width: maxDim, height: maxDim}}}
	viewportInf := vkPipelineViewportStateCreateInfo{
		sType:         _VK_STRUCTURE_TYPE_PIPELINE_VIEWPORT_STATE_CREATE_INFO,
		viewportCount: 1,
		scissorCount:  uint32(len(scissors)),
		pScissors:     &scissors[0],
	}
	enable := vkBool32(0)
	if blend {
		enable = 1
	}
	attBlendInf := vkPipelineColorBlendAttachmentState{
		blendEnable:         enable,
		srcColorBlendFactor: srcFactor,
		srcAlphaBlendFactor: srcFactor,
		dstColorBlendFactor: dstFactor,
		dstAlphaBlendFactor: dstFactor,
		colorBlendOp:        _VK_BLEND_OP_ADD,
		alphaBlendOp:        _VK_BLEND_OP_ADD,
		colorWriteMask:      _VK_COLOR_COMPONENT_R_BIT | _VK_COLOR_COMPONENT_G_BIT | _VK_COLOR_COMPONENT_B_BIT | _VK_COLOR_COMPONENT_A_BIT,
	}
	blendInf := vkPipelineColorBlendStateCreateInfo{
		sType:           _VK_STRUCTURE_TYPE_PIPELINE_COLOR_BLEND_STATE_CREATE_INFO,
		attachmentCount: 1,
		pAttachments:    &attBlendInf,
	}
	var vkBinds []vkVertexInputBindingDescription
	var vkAttrs []vkVertexInputAttributeDescription
	for _, b := range bindings {
		vkBinds = append(vkBinds, vkVertexInputBindingDescription{
			binding: uint32(b.Binding),
			stride:  uint32(b.Stride),
		})
	}
	for _, a := range attrs {
		vkAttrs = append(vkAttrs, vkVertexInputAttributeDescription{
			location: uint32(a.Location),
			binding:  uint32(a.Binding),
			format:   a.Format,
			offset:   uint32(a.Offset),
		})
	}
	vertexInf := vkPipelineVertexInputStateCreateInfo{
		sType: _VK_STRUCTURE_TYPE_PIPELINE_VERTEX_INPUT_STATE_CREATE_INFO,
	}
	if n := len(vkBinds); n > 0 {
		vertexInf.vertexBindingDescriptionCount = uint32(n)
		vertexInf.pVertexBindingDescriptions = &vkBinds[0]
	}
	if n := len(vkAttrs); n > 0 {
		vertexInf.vertexAttributeDescriptionCount = uint32(n)
		vertexInf.pVertexAttributeDescriptions = &vkAttrs[0]
	}
	inf := vkGraphicsPipelineCreateInfo{
		sType:      _VK_STRUCTURE_TYPE_GRAPHICS_PIPELINE_CREATE_INFO,
		stageCount: uint32(len(stages)),
		pStages:    &stages[0],
		renderPass: pass,
		layout:     layout,
		pRasterizationState: &vkPipelineRasterizationStateCreateInfo{
			sType:     _VK_STRUCTURE_TYPE_PIPELINE_RASTERIZATION_STATE_CREATE_INFO,
			lineWidth: 1.0,
		},
		pMultisampleState: &vkPipelineMultisampleStateCreateInfo{
			sType:                _VK_STRUCTURE_TYPE_PIPELINE_MULTISAMPLE_STATE_CREATE_INFO,
			rasterizationSamples: _VK_SAMPLE_COUNT_1_BIT,
		},
		pInputAssemblyState: &vkPipelineInputAssemblyStateCreateInfo{
			sType:    _VK_STRUCTURE_TYPE_PIPELINE_INPUT_ASSEMBLY_STATE_CREATE_INFO,
			topology: topology,
		},
	}

	var pipe vkPipeline
	if err := vkErr(vkCreateGraphicsPipelines(funcs.vkCreateGraphicsPipelines, d, nilPipelineCache, inf, dynInf, blendInf, vertexInf, viewportInf, nil, &pipe)); err != nil {
		return nilPipeline, fmt.Errorf("vulkan: vkCreateGraphicsPipelines: %w", err)
	}
	return pipe, nil
}

func DestroyPipeline(d Device, p Pipeline) {
	vkDestroyPipeline(funcs.vkDestroyPipeline, d, p, nil)
}

func CreatePipelineLayout(d Device, pushRanges []PushConstantRange, sets []DescriptorSetLayout) (PipelineLayout, error) {
	inf := vkPipelineLayoutCreateInfo{
		sType: _VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO,
	}
	if n := len(sets); n > 0 {
		inf.setLayoutCount = uint32(n)
		inf.pSetLayouts = &sets[0]
	}
	if n := len(pushRanges); n > 0 {
		inf.pushConstantRangeCount = uint32(n)
		inf.pPushConstantRanges = &pushRanges[0]
	}
	var l vkPipelineLayout
	if err := vkErr(vkCreatePipelineLayout(funcs.vkCreatePipelineLayout, d, inf, nil, &l)); err != nil {
		return nilPipelineLayout, fmt.Errorf("vulkan: vkCreatePipelineLayout: %w", err)
	}
	return l, nil
}

func DestroyPipelineLayout(d Device, l PipelineLayout) {
	vkDestroyPipelineLayout(funcs.vkDestroyPipelineLayout, d, l, nil)
}

func CreateDescriptorSetLayout(d Device, bindings []DescriptorSetLayoutBinding) (DescriptorSetLayout, error) {
	var vkbinds []vkDescriptorSetLayoutBinding
	for _, b := range bindings {
		vkbinds = append(vkbinds, vkDescriptorSetLayoutBinding{
			binding:         uint32(b.Binding),
			descriptorType:  b.DescriptorType,
			descriptorCount: 1,
			stageFlags:      b.StageFlags,
		})
	}
	inf := vkDescriptorSetLayoutCreateInfo{
		sType: _VK_STRUCTURE_TYPE_DESCRIPTOR_SET_LAYOUT_CREATE_INFO,
	}
	if n := len(vkbinds); n > 0 {
		inf.bindingCount = uint32(len(vkbinds))
		inf.pBindings = &vkbinds[0]
	}
	var l vkDescriptorSetLayout
	if err := vkErr(vkCreateDescriptorSetLayout(funcs.vkCreateDescriptorSetLayout, d, inf, nil, &l)); err != nil {
		return nilDescriptorSetLayout, fmt.Errorf("vulkan: vkCreateDescriptorSetLayout: %w", err)
	}
	return l, nil
}

func DestroyDescriptorSetLayout(d Device, l DescriptorSetLayout) {
	vkDestroyDescriptorSetLayout(funcs.vkDestroyDescriptorSetLayout, d, l, nil)
}

func MapMemory(d Device, mem DeviceMemory, offset, size int) ([]byte, error) {
	var ptr unsafe.Pointer
	if err := vkErr(vkMapMemory(funcs.vkMapMemory, d, mem, vkDeviceSize(offset), vkDeviceSize(size), 0, &ptr)); err != nil {
		return nil, fmt.Errorf("vulkan: vkMapMemory: %w", err)
	}
	return ((*[1 << 30]byte)(ptr))[:size:size], nil
}

func UnmapMemory(d Device, mem DeviceMemory) {
	vkUnmapMemory(funcs.vkUnmapMemory, d, mem)
}

func ResetCommandBuffer(buf CommandBuffer) error {
	if err := vkErr(vkResetCommandBuffer(funcs.vkResetCommandBuffer, buf, 0)); err != nil {
		return fmt.Errorf("vulkan: vkResetCommandBuffer. %w", err)
	}
	return nil
}

func CreateDescriptorPool(d Device, maxSets int, sizes []DescriptorPoolSize) (DescriptorPool, error) {
	inf := vkDescriptorPoolCreateInfo{
		sType:         _VK_STRUCTURE_TYPE_DESCRIPTOR_POOL_CREATE_INFO,
		maxSets:       uint32(maxSets),
		poolSizeCount: uint32(len(sizes)),
		pPoolSizes:    &sizes[0],
	}
	var pool vkDescriptorPool
	if err := vkErr(vkCreateDescriptorPool(funcs.vkCreateDescriptorPool, d, inf, nil, &pool)); err != nil {
		return nilDescriptorPool, fmt.Errorf("vulkan: vkCreateDescriptorPool: %w", err)
	}
	return pool, nil
}

func DestroyDescriptorPool(d Device, pool DescriptorPool) {
	vkDestroyDescriptorPool(funcs.vkDestroyDescriptorPool, d, pool, nil)
}

func ResetDescriptorPool(d Device, pool DescriptorPool) error {
	if err := vkErr(vkResetDescriptorPool(funcs.vkResetDescriptorPool, d, pool, 0)); err != nil {
		return fmt.Errorf("vulkan: vkResetDescriptorPool: %w", err)
	}
	return nil
}

func UpdateDescriptorSet(d Device, write WriteDescriptorSet) {
	vkUpdateDescriptorSets(funcs.vkUpdateDescriptorSets, d, write, 0, nil)
}

func AllocateDescriptorSets(d Device, pool DescriptorPool, layout DescriptorSetLayout, count int) ([]DescriptorSet, error) {
	layouts := make([]DescriptorSetLayout, count)
	for i := range layouts {
		layouts[i] = layout
	}
	inf := vkDescriptorSetAllocateInfo{
		sType:              _VK_STRUCTURE_TYPE_DESCRIPTOR_SET_ALLOCATE_INFO,
		descriptorPool:     pool,
		descriptorSetCount: uint32(count),
		pSetLayouts:        &layouts[0],
	}
	sets := make([]DescriptorSet, count)
	if err := vkErr(vkAllocateDescriptorSets(funcs.vkAllocateDescriptorSets, d, inf, &sets[0])); err != nil {
		return nil, fmt.Errorf("vulkan: vkAllocateDescriptorSets: %w", err)
	}
	return sets, nil
}

func CreateComputePipeline(d Device, mod ShaderModule, layout PipelineLayout) (Pipeline, error) {
	main := cString("main")
	defer cFree(unsafe.Pointer(main))
	inf := vkComputePipelineCreateInfo{
		sType: _VK_STRUCTURE_TYPE_COMPUTE_PIPELINE_CREATE_INFO,
		stage: vkPipelineShaderStageCreateInfo{
			sType:  _VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO,
			stage:  _VK_SHADER_STAGE_COMPUTE_BIT,
			module: mod,
			pName:  main,
		},
		layout: layout,
	}
	var pipe vkPipeline
	if err := vkErr(vkCreateComputePipelines(funcs.vkCreateComputePipelines, d, nilPipelineCache, 1, &inf, nil, &pipe)); err != nil {
		return nilPipeline, fmt.Errorf("vulkan: vkCreateComputePipelines: %w", err)
	}
	return pipe, nil
}

func CreateFence(d Device, flags int) (Fence, error) {
	inf := vkFenceCreateInfo{
		sType: _VK_STRUCTURE_TYPE_FENCE_CREATE_INFO,
		flags: vkFenceCreateFlags(flags),
	}
	var f vkFence
	if err := vkErr(vkCreateFence(funcs.vkCreateFence, d, &inf, nil, &f)); err != nil {
		return nilFence, fmt.Errorf("vulkan: vkCreateFence: %w", err)
	}
	return f, nil
}

func DestroyFence(d Device, f Fence) {
	vkDestroyFence(funcs.vkDestroyFence, d, f, nil)
}

func WaitForFences(d Device, fences ...Fence) error {
	if len(fences) == 0 {
		return nil
	}
	err := vkErr(vkWaitForFences(funcs.vkWaitForFences, d, uint32(len(fences)), &fences[0], _VK_TRUE, 0xffffffffffffffff))
	if err != nil {
		return fmt.Errorf("vulkan: vkWaitForFences: %w", err)
	}
	return nil
}

func ResetFences(d Device, fences ...Fence) error {
	if len(fences) == 0 {
		return nil
	}
	err := vkErr(vkResetFences(funcs.vkResetFences, d, uint32(len(fences)), &fences[0]))
	if err != nil {
		return fmt.Errorf("vulkan: vkResetFences: %w", err)
	}
	return nil
}

func BuildSubpassDependency(srcStage, dstStage PipelineStageFlags, srcMask, dstMask AccessFlags, flags DependencyFlags) SubpassDependency {
	return vkSubpassDependency{
		srcSubpass:      _VK_SUBPASS_EXTERNAL,
		srcStageMask:    srcStage,
		srcAccessMask:   srcMask,
		dstSubpass:      0,
		dstStageMask:    dstStage,
		dstAccessMask:   dstMask,
		dependencyFlags: flags,
	}
}

func BuildPushConstantRange(stages ShaderStageFlags, offset, size int) PushConstantRange {
	return vkPushConstantRange{
		stageFlags: stages,
		offset:     uint32(offset),
		size:       uint32(size),
	}
}

func BuildDescriptorPoolSize(typ DescriptorType, count int) DescriptorPoolSize {
	return vkDescriptorPoolSize{
		_type:           typ,
		descriptorCount: uint32(count),
	}
}

func BuildWriteDescriptorSetImage(set DescriptorSet, binding int, typ DescriptorType, sampler Sampler, view ImageView, layout ImageLayout) WriteDescriptorSet {
	return vkWriteDescriptorSet{
		sType:           _VK_STRUCTURE_TYPE_WRITE_DESCRIPTOR_SET,
		dstSet:          set,
		dstBinding:      uint32(binding),
		descriptorCount: 1,
		descriptorType:  typ,
		pImageInfo: &vkDescriptorImageInfo{
			sampler:     sampler,
			imageView:   view,
			imageLayout: layout,
		},
	}
}

func BuildWriteDescriptorSetBuffer(set DescriptorSet, binding int, typ DescriptorType, buf Buffer) WriteDescriptorSet {
	return vkWriteDescriptorSet{
		sType:           _VK_STRUCTURE_TYPE_WRITE_DESCRIPTOR_SET,
		dstSet:          set,
		dstBinding:      uint32(binding),
		descriptorCount: 1,
		descriptorType:  typ,
		pBufferInfo: &vkDescriptorBufferInfo{
			buffer: buf,
			_range: _VK_WHOLE_SIZE,
		},
	}
}

func PushConstantRangeStageFlags(r PushConstantRange) ShaderStageFlags {
	return r.stageFlags
}

func PushConstantRangeOffset(r PushConstantRange) int {
	return int(r.offset)
}

func PushConstantRangeSize(r PushConstantRange) int {
	return int(r.size)
}

func QueueFamilyPropertiesFlags(p QueueFamilyProperties) QueueFlags {
	return p.queueFlags
}

func SurfaceCapabilitiesMinExtent(c SurfaceCapabilities) image.Point {
	return image.Pt(int(c.minImageExtent.width), int(c.minImageExtent.height))
}

func SurfaceCapabilitiesMaxExtent(c SurfaceCapabilities) image.Point {
	return image.Pt(int(c.maxImageExtent.width), int(c.maxImageExtent.height))
}

func BuildViewport(x, y, width, height float32) Viewport {
	return vkViewport{
		x:        float32(x),
		y:        float32(y),
		width:    float32(width),
		height:   float32(height),
		maxDepth: 1.0,
	}
}

func BuildImageMemoryBarrier(img Image, srcMask, dstMask AccessFlags, oldLayout, newLayout ImageLayout, baseMip, numMips int) ImageMemoryBarrier {
	return vkImageMemoryBarrier{
		sType:         _VK_STRUCTURE_TYPE_IMAGE_MEMORY_BARRIER,
		srcAccessMask: srcMask,
		dstAccessMask: dstMask,
		oldLayout:     oldLayout,
		newLayout:     newLayout,
		image:         img,
		subresourceRange: vkImageSubresourceRange{
			aspectMask:   _VK_IMAGE_ASPECT_COLOR_BIT,
			baseMipLevel: uint32(baseMip),
			levelCount:   uint32(numMips),
			layerCount:   _VK_REMAINING_ARRAY_LAYERS,
		},
	}
}

func BuildBufferMemoryBarrier(buf Buffer, srcMask, dstMask AccessFlags) BufferMemoryBarrier {
	return vkBufferMemoryBarrier{
		sType:         _VK_STRUCTURE_TYPE_BUFFER_MEMORY_BARRIER,
		srcAccessMask: srcMask,
		dstAccessMask: dstMask,
		buffer:        buf,
		size:          _VK_WHOLE_SIZE,
	}
}

func BuildMemoryBarrier(srcMask, dstMask AccessFlags) MemoryBarrier {
	return vkMemoryBarrier{
		sType:         _VK_STRUCTURE_TYPE_MEMORY_BARRIER,
		srcAccessMask: srcMask,
		dstAccessMask: dstMask,
	}
}

func BuildBufferImageCopy(bufOff, bufStride, x, y, width, height int) BufferImageCopy {
	return vkBufferImageCopy{
		bufferOffset:    vkDeviceSize(bufOff),
		bufferRowLength: uint32(bufStride),
		imageSubresource: vkImageSubresourceLayers{
			aspectMask: _VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
		},
		imageOffset: vkOffset3D{
			x: int32(x), y: int32(y), z: 0,
		},
		imageExtent: vkExtent3D{
			width: uint32(width), height: uint32(height), depth: 1,
		},
	}
}

func BuildImageCopy(srcX, srcY, dstX, dstY, width, height int) ImageCopy {
	return vkImageCopy{
		srcSubresource: vkImageSubresourceLayers{
			aspectMask: _VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
		},
		srcOffset: vkOffset3D{
			x: int32(srcX),
			y: int32(srcY),
		},
		dstSubresource: vkImageSubresourceLayers{
			aspectMask: _VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
		},
		dstOffset: vkOffset3D{
			x: int32(dstX),
			y: int32(dstY),
		},
		extent: vkExtent3D{
			width:  uint32(width),
			height: uint32(height),
			depth:  1,
		},
	}
}

func BuildImageBlit(srcX, srcY, dstX, dstY, srcWidth, srcHeight, dstWidth, dstHeight, srcMip, dstMip int) ImageBlit {
	return vkImageBlit{
		srcOffsets: [2]vkOffset3D{
			{int32(srcX), int32(srcY), 0},
			{int32(srcWidth), int32(srcHeight), 1},
		},
		srcSubresource: vkImageSubresourceLayers{
			aspectMask: _VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
			mipLevel:   uint32(srcMip),
		},
		dstOffsets: [2]vkOffset3D{
			{int32(dstX), int32(dstY), 0},
			{int32(dstWidth), int32(dstHeight), 1},
		},
		dstSubresource: vkImageSubresourceLayers{
			aspectMask: _VK_IMAGE_ASPECT_COLOR_BIT,
			layerCount: 1,
			mipLevel:   uint32(dstMip),
		},
	}
}

func findMemoryTypeIndex(pd vkPhysicalDevice, constraints uint32, wantProps vkMemoryPropertyFlags) (int, bool) {
	var memProps vkPhysicalDeviceMemoryProperties
	vkGetPhysicalDeviceMemoryProperties(funcs.vkGetPhysicalDeviceMemoryProperties, pd, &memProps)

	for i := 0; i < int(memProps.memoryTypeCount); i++ {
		if (constraints & (1 << i)) == 0 {
			continue
		}
		if (memProps.memoryTypes[i].propertyFlags & wantProps) != wantProps {
			continue
		}
		return i, true
	}

	return 0, false
}

func choosePresentMode(pd vkPhysicalDevice, surf Surface) (vkPresentModeKHR, bool, error) {
	var count uint32
	err := vkErr(vkGetPhysicalDeviceSurfacePresentModesKHR(funcs.vkGetPhysicalDeviceSurfacePresentModesKHR, pd, surf, &count, nil))
	if err != nil {
		return 0, false, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfacePresentModesKHR: %w", err)
	}
	if count == 0 {
		return 0, false, nil
	}
	modes := make([]vkPresentModeKHR, count)
	err = vkErr(vkGetPhysicalDeviceSurfacePresentModesKHR(funcs.vkGetPhysicalDeviceSurfacePresentModesKHR, pd, surf, &count, &modes[0]))
	if err != nil {
		return 0, false, fmt.Errorf("vulkan: kGetPhysicalDeviceSurfacePresentModesKHR: %w", err)
	}
	for _, m := range modes {
		if m == _VK_PRESENT_MODE_MAILBOX_KHR || m == _VK_PRESENT_MODE_FIFO_KHR {
			return m, true, nil
		}
	}
	return 0, false, nil
}

func chooseFormat(pd vkPhysicalDevice, surf Surface) (vkSurfaceFormatKHR, bool, error) {
	var count uint32
	err := vkErr(vkGetPhysicalDeviceSurfaceFormatsKHR(funcs.vkGetPhysicalDeviceSurfaceFormatsKHR, pd, surf, &count, nil))
	if err != nil {
		return vkSurfaceFormatKHR{}, false, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceFormatsKHR: %w", err)
	}
	if count == 0 {
		return vkSurfaceFormatKHR{}, false, nil
	}
	formats := make([]vkSurfaceFormatKHR, count)
	err = vkErr(vkGetPhysicalDeviceSurfaceFormatsKHR(funcs.vkGetPhysicalDeviceSurfaceFormatsKHR, pd, surf, &count, &formats[0]))
	if err != nil {
		return vkSurfaceFormatKHR{}, false, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceFormatsKHR: %w", err)
	}
	// Query for format with sRGB support.
	// TODO: Support devices without sRGB.
	for _, f := range formats {
		if f.colorSpace != _VK_COLOR_SPACE_SRGB_NONLINEAR_KHR {
			continue
		}
		switch f.format {
		case _VK_FORMAT_B8G8R8A8_SRGB, _VK_FORMAT_R8G8B8A8_SRGB:
			return f, true, nil
		}
	}
	return vkSurfaceFormatKHR{}, false, nil
}

func chooseQueue(pd vkPhysicalDevice, surf Surface, flags vkQueueFlags) (int, bool, error) {
	queues := GetPhysicalDeviceQueueFamilyProperties(pd)
	for i, q := range queues {
		// Check for presentation and feature support.
		if q.queueFlags&flags != flags {
			continue
		}
		if surf != nilSurface {
			// Check for presentation support. It is possible that a device has no
			// queue with both rendering and presentation support, but not in reality.
			// See https://github.com/KhronosGroup/Vulkan-Docs/issues/1234.
			var support vkBool32
			if err := vkErr(vkGetPhysicalDeviceSurfaceSupportKHR(funcs.vkGetPhysicalDeviceSurfaceSupportKHR, pd, uint32(i), surf, &support)); err != nil {
				return 0, false, fmt.Errorf("vulkan: vkGetPhysicalDeviceSurfaceSupportKHR: %w", err)
			}
			if support != _VK_TRUE {
				continue
			}
		}
		return i, true, nil
	}
	return 0, false, nil
}

func vkErr(res vkResult) error {
	switch res {
	case _VK_SUCCESS:
		return nil
	default:
		return Error(res)
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("error %d", e)
}
