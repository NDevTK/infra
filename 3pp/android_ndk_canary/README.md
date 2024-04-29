Much of the scripts in this directory have been shamelessly stolen from the
Chromium parent in `chromium/src/third_party/android_toolchain/`.
The Chromium Android NDK does not contain a full toolchain, as per [CL 5106970](https://chromium-review.googlesource.com/c/chromium/src/+/5106970).

While this decision makes sense for the Chromium team, there are
dependencies (such as BoringSSL) which require a full NDK toolchain.
Thus concludes the origin story of the `infra/3pp/android_toolchain`.
