#include <stdlib.h>
#include <vips/vips.h>
#include <vips/vips7compat.h>


int
cgo_vips_init() {
    return vips_init("fotomat");
}
