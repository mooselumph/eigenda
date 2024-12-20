// SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.9;

import "../../src/libraries/EigenDARollupUtils.sol";
import "forge-std/Test.sol";

contract EigenDABlobUtilsHarness is Test {    

    function verifyBlob(
        IEigenDAServiceManager.BlobHeader calldata blobHeader,
        IEigenDAServiceManager eigenDAServiceManager,
        EigenDARollupUtils.BlobVerificationProof calldata blobVerificationProof
    ) external view {
        EigenDARollupUtils.verifyBlob(blobHeader, eigenDAServiceManager, blobVerificationProof);
    }

    function verifyBlobs(
        IEigenDAServiceManager.BlobHeader[] calldata blobHeaders,
        IEigenDAServiceManager eigenDAServiceManager,
        EigenDARollupUtils.BlobVerificationProof[] calldata blobVerificationProofs
    ) external view {
        EigenDARollupUtils.verifyBlobs(blobHeaders, eigenDAServiceManager, blobVerificationProofs);
    }
}
